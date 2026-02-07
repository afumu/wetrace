package lz4

import (
	"github.com/pierrec/lz4/v4"
)

func Decompress(src []byte) ([]byte, error) {
	// FIXME: lz4 的压缩率预计不到 3，这里设置了 4 保险一点
	out := make([]byte, len(src)*4)

	n, err := lz4.UncompressBlock(src, out)
	if err != nil {
		return nil, err
	}
	return out[:n], nil
}
