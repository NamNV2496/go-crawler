package entity

import (
	"bytes"
	"encoding/json"
	"text/template"
)

type GoldResponse struct {
	Data    []GoldData `json:"Data"`
	Message *string    `json:"Message"`
	Success bool       `json:"Success"`
}

type GoldData struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	BuyPrice        float64 `json:"buyPrice"`
	BuyChangePrice  float64 `json:"buyChangePrice"`
	SellPrice       float64 `json:"sellPrice"`
	SellChangePrice float64 `json:"sellChangePrice"`
	Zone            string  `json:"zone"`
	LastUpdated     string  `json:"lastUpdated"`
}

var GoldTemplate = `<b>Gold Price</b>:{{range .Data}}
	Tên: {{.Name}}, giá mua: {{.BuyPrice}}, giá bán: {{.SellPrice}}{{end}}`

func ExtractGoldPrice(data []byte) string {
	var response GoldResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return ""
	}

	goldTemplate, err := template.New("gold").Parse(GoldTemplate)
	if err != nil {
		return ""
	}

	var output bytes.Buffer
	err = goldTemplate.Execute(&output, response)
	if err != nil {
		return ""
	}
	resp := output.String()
	return resp
}
