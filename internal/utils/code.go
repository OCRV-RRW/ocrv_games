package utils

import (
	"github.com/jgweir/randstring"
)

func GenerateCode(length int) string {
	rs := randstring.New(length)
	rs.Specials(false)
	verificationCode, _ := rs.Build()
	return verificationCode
}
