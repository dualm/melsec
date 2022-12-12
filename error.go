package melsec

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errorTimeout = errors.New("以太网模块和PLC CPU之间的通讯时间超过CPU监视定时器的时间")
	errorUnknown = errors.New("未知错误")
)

func ErrorSelect(errCode []byte) error {
	err := errorUnknown

	switch {
	case reflect.DeepEqual(errCode, []byte{0x5E, 0xC0}):
		err = errorTimeout
	}

	return fmt.Errorf("%w, error code: %x", err, errCode)
}
