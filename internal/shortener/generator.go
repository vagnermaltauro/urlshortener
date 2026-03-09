package shortener

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/speps/go-hashids/v2"
)

func Generate() string {

	timestamp := time.Now().UnixNano()
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(999999))

	hd := hashids.NewData()
	hd.Salt = "super-segredo-2025-change-in-prod"
	hd.MinLength = 6
	h, _ := hashids.NewWithData(hd)

	code, _ := h.Encode([]int{int(timestamp % 1000000000), int(randomNum.Int64())})
	return code
}
