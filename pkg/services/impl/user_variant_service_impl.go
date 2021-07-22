package impl

import (
	"errors"
	"flagsense-go-sdk/pkg/dto"
	"flagsense-go-sdk/pkg/util"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/teltech/logger"
	"github.com/twmb/murmur3"
	"math"
	"strings"
)

const (
	TOTAL_THREE_DECIMAL_TRAFFIC = 100000
)

var MAX_HASH_VALUE = math.Pow(2, 32)

type UserVariantServiceImpl struct {
	Data   *dto.Data
	logger *logger.Log
}

func NewUserVariantService(data *dto.Data, logger *logger.Log) *UserVariantServiceImpl {
	return &UserVariantServiceImpl{
		Data:   data,
		logger: logger,
	}
}

func (uvs *UserVariantServiceImpl) GetUserVariant(userVariantDTO *dto.UserVariantDTO) error {
	if strings.TrimSpace(userVariantDTO.UserId) == "" {
		return errors.New(fmt.Sprintf("Bad user: %s", userVariantDTO.UserId))
	}
	flagDTO := uvs.getFlagData(userVariantDTO.FlagId)

	// assumption ID is always present for flag
	if flagDTO.ID == "" {
		return errors.New("Flag not found")
	}

	if flagDTO.Type != userVariantDTO.ExpectedVariantType {
		return errors.New("Bad flag type specified")
	}

	userVariantKey := uvs.getUserVariantKey(*userVariantDTO, flagDTO)
	userVariantDTO.Key = userVariantKey
	userVariantDTO.Value = flagDTO.Variants[userVariantKey].Value

	return nil
}

func (uvs *UserVariantServiceImpl) getUserVariantKey(userVariantDTO dto.UserVariantDTO, flagDTO dto.FlagDTO) string {
	userId := userVariantDTO.UserId
	attributes := userVariantDTO.Attributes

	envData := flagDTO.EnvData
	if envData.Status == dto.INACTIVE {
		return envData.OffVariant
	}

	segments := uvs.getSegmentsMap()
	if !uvs.matchesPrerequisites(userId, attributes, envData.PreRequisites, segments) {
		return envData.OffVariant
	}

	targetUsers := envData.TargetUsers
	if targetUsers != nil && targetUsers[userId] != "" {
		return targetUsers[userId]
	}

	targetSegmentsOrder := envData.TargetSegmentsOrder
	if targetSegmentsOrder != nil {
		for _, targetSegment := range targetSegmentsOrder {
			if uvs.isUserInSegment(userId, attributes, segments[targetSegment]) {
				return uvs.allocateTrafficVariant(userId, flagDTO, envData.TargetSegments[targetSegment])
			}
		}
	}
	return uvs.allocateTrafficVariant(userId, flagDTO, envData.Traffic)
}

func (uvs *UserVariantServiceImpl) getFlagData(flagId string) dto.FlagDTO {
	if uvs.Data.Flags == nil || len(uvs.Data.Flags) == 0 {
		return dto.FlagDTO{}
	}
	return uvs.Data.Flags[flagId]
}

func (uvs *UserVariantServiceImpl) getSegmentsMap() map[string]dto.SegmentDTO {
	if uvs.Data.Segments == nil || len(uvs.Data.Segments) == 0 {
		return map[string]dto.SegmentDTO{}
	}
	return uvs.Data.Segments
}

func (uvs *UserVariantServiceImpl) matchesPrerequisites(userId string, attributes map[string]interface{},
	prerequisites []string, segmentsMap map[string]dto.SegmentDTO) bool {

	if prerequisites == nil || len(prerequisites) == 0 {
		return true
	}

	for _, prerequisite := range prerequisites {
		if !uvs.isUserInSegment(userId, attributes, segmentsMap[prerequisite]) {
			return false
		}
	}
	return true
}

func (uvs *UserVariantServiceImpl) isUserInSegment(userId string, attributes map[string]interface{}, segmentDTO dto.SegmentDTO) bool {
	// assuming that ID is mandatory
	if segmentDTO.ID == "" {
		return false
	}

	for _, rule := range segmentDTO.Rules {
		if !uvs.matchesAndRule(userId, attributes, rule) {
			return false
		}
	}
	return true
}

func (uvs *UserVariantServiceImpl) matchesAndRule(userId string, attributes map[string]interface{}, orRules []*dto.RulesDTO) bool {
	for _, orRule := range orRules {
		if uvs.matchesRule(userId, attributes, orRule) {
			return true
		}
	}
	return false
}

func (uvs *UserVariantServiceImpl) matchesRule(userId string, attributes map[string]interface{}, rule *dto.RulesDTO) bool {
	attributeValue := uvs.getAttributeValue(userId, attributes, rule.Key)
	if attributeValue == nil {
		return false
	}
	attributeValueType := fmt.Sprintf("%T", attributeValue)
	var userMatchesRule bool

	switch rule.Type {
	case dto.INT:
		if !strings.Contains(attributeValueType, "int") && !strings.Contains(attributeValueType, "float") {
			return false
		}
		userMatchesRule = uvs.matchesInt32Rule(rule, util.ConvertToInt32(attributeValue))
		break
	case dto.BOOL:
		if attributeValueType != "bool" {
			return false
		}
		userMatchesRule = uvs.matchesBoolRule(rule, attributeValue.(bool))
		break
	case dto.DOUBLE:
		if !strings.Contains(attributeValueType, "float") && !strings.Contains(attributeValueType, "int") {
			return false
		}
		userMatchesRule = uvs.matchesFloat64Rule(rule, util.ConvertToFloat64(attributeValue))
		break
	case dto.STRING:
		if attributeValueType != "string" {
			return false
		}
		userMatchesRule = uvs.matchesStringRule(rule, attributeValue.(string))
		break
	case dto.VERSION:
		if attributeValueType != "string" {
			return false
		}
		userMatchesRule = uvs.matchesVersionRule(rule, attributeValue.(string))
		break
	default:
		userMatchesRule = false
	}

	return userMatchesRule == rule.Match
}

func (uvs *UserVariantServiceImpl) getAttributeValue(userId string, attributes map[string]interface{}, key string) interface{} {
	if key == "id" {
		return userId
	}
	if attributes == nil {
		return nil
	}
	return attributes[key]

}

func (uvs *UserVariantServiceImpl) matchesFloat64Rule(rule *dto.RulesDTO, attributeValue float64) bool {
	values := rule.Values

	switch rule.Operator {
	case dto.LT:
		return attributeValue < values[0].(float64)
	case dto.LTE:
		return attributeValue <= values[0].(float64)
	case dto.EQ:
		return attributeValue == values[0].(float64)
	case dto.GT:
		return attributeValue > values[0].(float64)
	case dto.GTE:
		return attributeValue >= values[0].(float64)
	case dto.IOF:
		return util.Contains(values, attributeValue)
	default:
		return false
	}
}

func (uvs *UserVariantServiceImpl) matchesInt32Rule(rule *dto.RulesDTO, attributeValue int32) bool {
	values := rule.Values

	switch rule.Operator {
	case dto.LT:
		return attributeValue < values[0].(int32)
	case dto.LTE:
		return attributeValue <= values[0].(int32)
	case dto.EQ:
		return attributeValue == values[0].(int32)
	case dto.GT:
		return attributeValue > values[0].(int32)
	case dto.GTE:
		return attributeValue >= values[0].(int32)
	case dto.IOF:
		return util.Contains(values, attributeValue)
	default:
		return false
	}
}

func (uvs *UserVariantServiceImpl) matchesStringRule(rule *dto.RulesDTO, attributeValue string) bool {
	values := rule.Values
	switch rule.Operator {
	case dto.EQ:
		return attributeValue == values[0].(string)
	case dto.HAS:
		return strings.Contains(attributeValue, values[0].(string))
	case dto.SW:
		return strings.HasPrefix(attributeValue, values[0].(string))
	case dto.EW:
		return strings.HasSuffix(attributeValue, values[0].(string))
	case dto.IOF:
		return util.Contains(values, attributeValue)
	default:
		return false
	}
}

func (uvs *UserVariantServiceImpl) matchesBoolRule(rule *dto.RulesDTO, attributeValue bool) bool {
	values := rule.Values

	if rule.Operator == dto.EQ {
		return attributeValue == values[0]
	}
	return false
}

func (uvs *UserVariantServiceImpl) matchesVersionRule(rule *dto.RulesDTO, attributeValue string) bool {
	values := rule.Values
	attrVersion, err := version.NewVersion(attributeValue)
	if err != nil {
		attrVersion, _ = version.NewVersion("0.0")
	}
	valueVersion, _ := version.NewVersion(values[0].(string))

	switch rule.Operator {
	case dto.LT:
		return attrVersion.Compare(valueVersion) < 0
	case dto.LTE:
		return attrVersion.Compare(valueVersion) <= 0
	case dto.EQ:
		return attrVersion.Compare(valueVersion) == 0
	case dto.GT:
		return attrVersion.Compare(valueVersion) > 0
	case dto.GTE:
		return attrVersion.Compare(valueVersion) >= 0
	case dto.IOF:
		return util.ContainsVersion(values, attrVersion)
	default:
		return false
	}
}

func (uvs *UserVariantServiceImpl) allocateTrafficVariant(userId string, flagDTO dto.FlagDTO, traffic map[string]int) string {
	if len(traffic) == 1 {
		for key, _ := range traffic {
			return key
		}
	}

	bucketingId := userId + flagDTO.ID
	variantsOrder := flagDTO.VariantsOrder

	hasher := murmur3.SeedNew32(flagDTO.Seed)
	if _, err := hasher.Write([]byte(bucketingId)); err != nil {
		//uvs.logger.Error(fmt.Sprintf("error while generating hash for the bucket key=%s, err:%+v", bucketingId, err))
	}
	hashCode := hasher.Sum32()
	ratio := float64(hashCode) / MAX_HASH_VALUE
	bucketValue := int(TOTAL_THREE_DECIMAL_TRAFFIC * ratio)

	endOfRange := 0
	for _, variant := range variantsOrder {
		endOfRange += traffic[variant]
		if bucketValue < endOfRange {
			return variant
		}
	}

	return variantsOrder[len(variantsOrder)-1]
}
