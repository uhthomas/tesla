package tesla

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"io"
	"strconv"
	"strings"
)

func userAgent() string {
	const prefixBytes = 9
	var buf [prefixBytes + 8]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		// I know.
		panic(err)
	}

	x, _ := binary.Uvarint(buf[prefixBytes:])

	var b strings.Builder
	b.WriteString(base64.RawURLEncoding.EncodeToString(buf[:prefixBytes]))
	b.WriteRune('/')
	b.Write(strconv.AppendUint(nil, x, 10))
	return b.String()
}
