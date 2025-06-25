package urlbuilderfactory

// type CityBuilder struct {
// 	Db *gorm.DB
// }

// func NewCityBuilder(
// 	db *gorm.DB,
// ) *CityBuilder {
// 	return &CityBuilder{
// 		Db: db,
// 	}
// }

// var _ IBuilder = &CityBuilder{}

// func (_self *CityBuilder) Build(ctx context.Context, request map[string]string) ([]*entity.ShortLink, error) {
// 	resp := []*entity.ShortLink{}
// 	err := _self.Db.Model(&domain.ShortLink{}).Where("filter ->> 'city' = ?", request["city"]).Offset(0).Limit(5).Find(&resp).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func (_self *CityBuilder) BuildRecommend(ctx context.Context, request map[string]string, fields []QueryOption) ([]*entity.ShortLink, error) {
// 	var resp []*entity.ShortLink
// 	var data []*domain.ShortLink
// 	// TBU: use AI to recommend next cities
// 	// var nextCities []*entity.ShortLink
// 	city := request["city"]
// 	if city == "" {
// 		return nil, nil
// 	}
// 	tx := _self.Db.Model(&domain.ShortLink{})
// 	for _, field := range fields {
// 		if field.And {
// 			tx = tx.Where("filter->>'"+field.Field+"' =?", request[field.Field])
// 		} else {
// 			tx = tx.Or("filter->>'"+field.Field+"' =?", request[field.Field])
// 		}
// 	}
// 	if err := tx.
// 		Offset(0).
// 		Limit(5).
// 		Find(&data).Error; err != nil {
// 		return nil, err
// 	}

// 	utils.Copy(&resp, data)
// 	return resp, nil
// }
