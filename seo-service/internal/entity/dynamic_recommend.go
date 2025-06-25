package entity

type DynamicRecommend struct {
	Data []*DynamicRecommendGroup `json:"data"`
}

type DynamicRecommendGroup struct {
	Group string       `json:"group"`
	Data  []*ShortLink `json:"data"`
}
