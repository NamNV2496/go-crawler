package utils

import (
	"github.com/jinzhu/copier"
)

func Copy(dst any, src any) error {
	return copier.Copy(dst, src)
}
