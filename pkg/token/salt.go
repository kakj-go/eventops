package token

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
)

func Salt() string {
	randValue := rand.Intn(1000000)
	Md5Inst := md5.New()
	Md5Inst.Write([]byte(strconv.FormatInt(int64(randValue), 10)))
	return hex.EncodeToString(Md5Inst.Sum(nil))
}
