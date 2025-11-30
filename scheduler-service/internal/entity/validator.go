package entity

type FieldValidation struct {
	Label     string `json:"label,omitempty"`
	MinLength int    `json:"min_length,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	MinValue  int    `json:"min_value,omitempty"`
	MaxValue  int    `json:"max_value,omitempty"`
	MinWord   int    `json:"min_word,omitempty"`
	MaxWord   int    `json:"max_word,omitempty"`
}

type Requires struct {
	Require map[string]map[string]FieldRequire `json:"label,omitempty"`
}

type FieldRequire map[string]string

type FieldRequireCondition struct {
	Conditions []string
	Param      string
	Name       string
}
