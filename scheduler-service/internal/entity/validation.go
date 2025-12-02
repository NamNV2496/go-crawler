package entity

const (
	OP_GT  = "gt"
	OP_LT  = "lt"
	OP_GTE = "gte"
	OP_LTE = "lte"
	OP_EQ  = "eq"
	OP_NE  = "ne"
)

type CrossFieldRule struct {
	Pattern       string
	ErrorMsg      string
	AllowedValues []string
	Operator      string
	Field         string
	Value         string
}
