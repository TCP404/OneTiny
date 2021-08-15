package util

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"
)

func RandString(n int) string {
	var letters = []byte("!qaz@wsQWERTx#edYUIOPc$rfvtASDFGgb^yHJKLhnZXCVB&uNMjm*ik,(ol.)p;/_['+]")
	result := make([]byte, n)
	lettersLen := int64(len(letters))
	rand.Seed(time.Now().Unix())
	for i := range result {
		result[i] = letters[rand.Int63()%lettersLen]
	}
	return string(result)
}

func MD5(content string) string {
	h := md5.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}
