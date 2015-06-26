package util

import (
	"crypto/md5"
)

func Md5Encrypt(in string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(in))
	outStr := md5Ctx.Sum(nil)
	return string(outStr)
}
