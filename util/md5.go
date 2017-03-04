package util

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5Encrypt(in string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(in))
	bytes := md5Ctx.Sum(nil)
	return string(bytes)
}

func HexMd5Encrypt(in string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(in))
	bytes := md5Ctx.Sum(nil)
	return hex.EncodeToString(bytes)
}
