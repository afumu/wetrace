package zstd

import (
	"github.com/klauspost/compress/zstd"
)

var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))

func Decompress(src []byte) ([]byte, error) {
	return decoder.DecodeAll(src, nil)
}
