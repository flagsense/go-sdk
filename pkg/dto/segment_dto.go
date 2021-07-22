package dto

type SegmentDTO struct {
	ID    string        `json:"id"`
	Rules [][]*RulesDTO `json:"rules"`
}

type RulesDTO struct {
	Key      string        `json:"key"`
	Match    bool          `json:"match"`
	Operator string        `json:"operator"`
	Type     string        `json:"type"`
	Values   []interface{} `json:"values"`
}

func (sd *SegmentDTO) MatchType() {
	for _, rules := range sd.Rules {
		for _, rule := range rules {
			rule.MatchType()
		}
	}
}

func (rd *RulesDTO) MatchType() {
	if rd.Type == INT || rd.Type == INT32 || rd.Type == INT64 {
		values := make([]interface{}, len(rd.Values))
		for i, val := range rd.Values {
			switch val.(type) {
			case float64:
				values[i] = int32(val.(float64))
			case float32:
				values[i] = int32(val.(float32))
			case int32:
				values[i] = int32(val.(int32))
			case int64:
				values[i] = int32(val.(int64))
			default:
				values[i] = int32(val.(float64))
			}
		}
		rd.Values = values
	} else if rd.Type == FLOAT64 || rd.Type == FLOAT32 || rd.Type == FLOAT {
		values := make([]interface{}, len(rd.Values))
		for i, val := range rd.Values {
			switch val.(type) {
			case float32:
				values[i] = float64(val.(float32))
			case int32:
				values[i] = float64(val.(int32))
			case int64:
				values[i] = float64(val.(int64))
			case int:
				values[i] = float64(val.(int))
			default:
				values[i] = val.(float64)
			}
		}
		rd.Values = values
	}
}
