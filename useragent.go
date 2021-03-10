package tesla

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"io"
	"strconv"
	"strings"
)

func userAgent() string {
	const prefixBytes = 6
	var buf [prefixBytes + 8]byte
	if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
		// I know.
		panic(err)
	}

	var b strings.Builder
	e := base32.NewEncoder(base32.StdEncoding.WithPadding(base32.NoPadding), &b)
	e.Write(buf[:prefixBytes])
	e.Close()
	b.WriteRune('/')
	b.Write(strconv.AppendUint(nil, binary.BigEndian.Uint64(buf[prefixBytes:]), 10))
	return b.String()
}
