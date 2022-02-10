package melsec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// 编码软元件
func encodeSoftComponent(component string) (McMessage, error) {
	name, no := splitComponetName(component)

	encodeName, base := encodeComponentName(name)

	if encodeName == nil && base == -1 {
		return nil, errors.New("wrong component name")
	}

	n, err := strconv.ParseUint(no, base, 64)
	if err != nil {
		return nil, err
	}

	// Q系列3字节软元件编号
	// iQ系列4字节软件编号
	offset, err := encodeUint(n, 3)
	if err != nil {
		return nil, err
	}

	return append(offset, encodeName...), nil
}

// fix, eg: XA0
func splitComponetName(component string) (string, string) {
	index := strings.IndexFunc(component, func(r rune) bool {
		return unicode.IsDigit(r)
	})

	return component[:index], component[index:]
}

func encodeComponentName(componentName string) (McMessage, int) {
	switch strings.ToLower(componentName) {
	case "m":
		return M_Component, Base_10
	case "x":
		return X_Component, Base_16
	case "w":
		return W_Component, Base_16
	case "d":
		return D_Component, Base_10
	case "r":
		return R_Component, Base_10
	case "b":
		return B_Component, Base_16
	default:
		return nil, -1
	}
}

// 返回一个软元件头是字还是位
// bit: 1, 0
// word: 0, 1
func componentBitSize(componentName string) (int8, int8) {
	switch strings.ToLower(componentName) {
	case "m", "x", "y", "b":
		return 1, 0
	case "d", "w", "r":
		return 0, 1
	}

	return 0, 0
}

// MELSEC iQ-R系列：4字节
// MELSEC Q系列：3字节
func encodeUint(num uint64, count int) (McMessage, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		return nil, err
	}

	return buf.Bytes()[:count], nil
}
