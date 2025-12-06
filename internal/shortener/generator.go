package shortener

import (
    "sync/atomic"
    "github.com/speps/go-hashids/v2"
)

var counter uint64

func Generate() string {
    id := atomic.AddUint64(&counter, 1)
    hd := hashids.NewData()
    hd.Salt = "super-segredo-2025-change-in-prod"
    hd.MinLength = 6
    h, _ := hashids.NewWithData(hd)
    code, _ := h.Encode([]int{int(id)})
    return code
}
