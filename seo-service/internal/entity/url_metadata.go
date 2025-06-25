package entity

type UrlMetadata struct {
	Id      int64  `json:"id,omitempty"`
	UrlId   int64  `json:"url_id,omitempty"`
	Keyword string `json:"keyword,omitempty"`
	Value   string `json:"value,omitempty"`
}
