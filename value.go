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
	name, no := splitComponentName(component)

	encodeName, base := encodeComponentName(name)

	if encodeName == nil && base == -1 {
		return nil, errors.New("wrong component name")
	}

	n, err := strconv.ParseUint(no, base, 64)
	if err != nil {
		return nil, err
	}

	// Q系列3个字节软元件编号
	// iQ系列4个字节软件编号
	offset, err := encodeUint(n, 3)
	if err != nil {
		return nil, err
	}

	return append(offset, encodeName...), nil
}

func splitComponentName(component string) (string, string) {
	component = strings.ToUpper(component)
	index := strings.IndexFunc(component, func(r rune) bool {
		return unicode.IsDigit(r)
	})

LOOP:
	for {
		switch index {
		case 1:
			switch component[:index] {
			case "X", "Y", "M", "L", "F", "V", "B", "D", "W", "Z", "R", "U":
				break LOOP
			default:
				return "", ""
			}
		case 2:
			switch component[:index] {
			case "SM", "SD", "TS", "TC", "TN", "CS", "CC", "CN", "SB", "SW", "DX", "DY", "LZ", "ZR", "RD":
				break LOOP
			default:
				index -= 1
			}
		case 3:
			switch component[:index] {
			case "LTS", "LTC", "LTN", "STS", "STC", "STN", "LCS", "LCC", "LCN":
				break LOOP
			default:
				index -= 1
			}
		case 4:
			switch component[:index] {
			case "LSTS", "LSTC", "LSTN":
				break LOOP
			default:
				index -= 1
			}
		default:
			index -= 1
		}
	}

	return component[:index], component[index:]
}

// todo, encodeComponentName, melsec通信协议参考手册 P66
func encodeComponentName(componentName string) (McMessage, int) {
	switch strings.ToLower(componentName) {
	case "m":
		return MComponent, Base10
	case "x":
		return XComponent, Base16
	case "w":
		return WComponent, Base16
	case "d":
		return DComponent, Base10
	case "r":
		return RComponent, Base10
	case "b":
		return BComponent, Base16
	case "sm":
		return SMComponent, Base10
	case "y":
		return YComponent, Base16
	case "sd":
		return SDComponent, Base10
	case "l":
		return LComponent, Base10
	case "f":
		return FComponent, Base10
	case "v":
		return VComponent, Base10
	case "tn":
		return TNComponent, Base10
	case "ts":
		return TSComponent, Base10
	case "tc":
		return TCComponent, Base10
	case "cn":
		return CNComponent, Base10
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
