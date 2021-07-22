package dto

type Data struct {
	Segments      map[string]SegmentDTO `json:"segments"`
	Flags         map[string]FlagDTO    `json:"flags"`
	LastUpdatedOn float64               `json:"lastUpdatedOn"`
}

func (d *Data) MatchType() {
	for _, value := range d.Segments {
		value.MatchType()
	}
}
