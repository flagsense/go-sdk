package dto

var VariantType = []string{
	"INT",
	"BOOL",
	"DOUBLE",
	"STRING",
	"JSON",
}

var Status = []string{
	"ACTIVE",
	"INACTIVE",
}

const (
	ACTIVE   = "ACTIVE"
	INACTIVE = "INACTIVE"

	INT32   = "INT32"
	INT     = "INT"
	FLOAT   = "FLOAT"
	FLOAT32 = "FLOAT32"
	INT64   = "INT64"
	BOOL    = "BOOL"
	FLOAT64 = "FLOAT64"
	STRING  = "STRING"
	JSON    = "JSON"
	MAP     = "MAP"
	DOUBLE  = "DOUBLE"
	VERSION = "VERSION"

	LT  = "LT"
	LTE = "LTE"
	EQ  = "EQ"
	GT  = "GT"
	GTE = "GTE"
	IOF = "IOF"

	HAS = "HAS"
	SW  = "SW"
	EW  = "EW"
)

type FlagDTO struct {
	ID            string             `json:"id"`
	Seed          uint32             `json:"seed"`
	Variants      map[string]Variant `json:"variants"`
	VariantsOrder []string           `json:"variantsOrder"`
	Type          string             `json:"type"`
	EnvData       EnvData            `json:"envData"`
}

type Variant struct {
	Value interface{} `json:"value"`
	Name  string      `json:"name"`
}

type EnvData struct {
	PreRequisites       []string                  `json:"preRequisites"`
	OffVariant          string                    `json:"offVariant"`
	TargetUsers         map[string]string         `json:"targetUsers"`
	TargetSegments      map[string]map[string]int `json:"targetSegments"`
	TargetSegmentsOrder []string                  `json:"targetSegmentsOrder"`
	Traffic             map[string]int            `json:"traffic"`
	Status              string                    `json:"status"`
}
