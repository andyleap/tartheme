// tartheme project tartheme.go
package tartheme

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"time"
)

const (
	blockSize = 512
	typeReg   = '0'
	typeRegA  = '\x00'
)

var zeroBlock = make([]byte, blockSize)
var errZeroBlock = errors.New("Zero Block")

func (tt *TarTheme) readAsset(pos int64) (*Asset, int64, error) {
	if pos+blockSize > int64(len(tt.tar)) {
		return nil, 0, io.EOF
	}
	header := tt.tar[pos : pos+blockSize]
	if bytes.Equal(header, zeroBlock) {
		return nil, pos + blockSize, errZeroBlock
	}

	asset := &Asset{}
	asset.Name = string(bytes.Trim(header[0:100], "\x00"))
	size, err := octal(header[124:136])
	if err != nil {
		return nil, pos + blockSize, err
	}
	modtime, err := octal(header[136:148])
	if err != nil {
		return nil, pos + blockSize, err
	}
	if header[156] != typeReg && header[156] != typeRegA {
		return nil, pos + blockSize + size + ((-size) & (blockSize - 1)), nil
	}
	asset.ModTime = time.Unix(modtime, 0)
	asset.Data = tt.tar[pos+blockSize : pos+blockSize+size]
	return asset, pos + blockSize + size + ((-size) & (blockSize - 1)), nil
}

func (tt *TarTheme) readAllAssets() error {
	pos := int64(0)
	var asset *Asset
	var err error
	countZero := 0
	for {
		asset, pos, err = tt.readAsset(pos)
		if err == errZeroBlock {
			countZero = countZero + 1
			if countZero == 2 {
				return nil
			}
			continue
		}
		if err != nil {
			return err
		}
		if asset != nil {
			tt.Assets[asset.Name] = asset
		}
	}
}

func octal(b []byte) (int64, error) {
	if len(b) > 0 && b[0]&0x80 != 0 {
		var x int64
		for i, c := range b {
			if i == 0 {
				c &= 0x7f
			}
			x = x<<8 | int64(c)
		}
		return x, nil
	}
	b = bytes.Trim(b, " \x00")
	if len(b) == 0 {
		return 0, nil
	}
	x, err := strconv.ParseUint(string(b), 8, 64)
	if err != nil {
		return 0, err
	}
	return int64(x), nil
}
