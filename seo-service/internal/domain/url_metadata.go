package domain

type UrlMetadata struct {
	Id      int64  `gorm:"column:id;primaryKey;unique" json:"id"`
	UrlId   int64  `gorm:"column:url_id" json:"url_id"`
	Keyword string `gorm:"column:keyword" json:"keyword"`
	Value   string `gorm:"column:value" json:"value"`
}

func (_self UrlMetadata) TableName() string {
	return "url_metadata"
}
