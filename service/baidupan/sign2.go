package baidupan

import (
	"bytes"
	"encoding/base64"
)

func realSign2(j, r []rune) []byte {
	var (
		a  = make([]rune, 256)
		p  = make([]rune, 256)
		o  = make([]byte, len(r))
		v  = len(j)
		q  int
		u  rune
		i  int
		k  rune
		dr int
	)
	if v == 0 {
		return o
	}
	for ; q < 256; q++ {
		dr = q % v
		a[q] = j[dr : 1+dr][0]
		p[q] = rune(q)
	}
	for q = 0; q < 256; q++ {
		u = (u + p[q] + a[q]) % 256
		p[q], p[u] = p[u], p[q]
	}
	u = 0
	for q = 0; q < len(r); q++ {
		i = (i + 1) % 256
		u = (u + p[i]) % 256
		p[i], p[u] = p[u], p[i]
		k = p[(p[i]+p[u])%256]
		o[q] = byte(r[q] ^ k)
	}
	return o
}

func base64Encode(raw []byte) []byte {
	var encoded bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
	_, _ = encoder.Write(raw)
	_ = encoder.Close()
	return encoded.Bytes()
}

func panSign2(sign1, sign3 string) string {
	o := realSign2([]rune(sign3), []rune(sign1))
	signed := base64Encode(o)
	return string(signed)
}
