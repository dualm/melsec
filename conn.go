package melsec

import (
	"bytes"
	"io"
	"net"
	"reflect"
)

const (
	FirstResponseLength     = 11 // 副帧头=2, 访问路径=7, 数据长度=2
	ResponseErrorCodeIndex  = 9
	ResponseErrorCodeLength = 2
)

func NewConn(addr, port string, ops ...PlcOption) (*PlcConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+port)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	if err = conn.SetKeepAlive(true); err != nil {
		return nil, err
	}

	return &PlcConn{
		conn,
		newPlcOption(ops),
	}, nil
}

type PlcConn struct {
	// conn   net.Conn
	net.Conn
	option *plcOptions
}

func (plc *PlcConn) SendCmd(msg McMessage, retSize int) ([]byte, error) {
	_, err := plc.Write(msg)
	if err != nil {
		return nil, err
	}

	buff := make([]byte, FirstResponseLength)

	_, err = io.ReadFull(plc, buff)
	if err != nil {
		return nil, err
	}

	// 返回错误代码
	if errorCode := buff[ResponseErrorCodeIndex : ResponseErrorCodeIndex+ResponseErrorCodeLength]; !reflect.DeepEqual(errorCode, CodeOK) {
		buff = make([]byte, ResponseErrorCodeLength)

		_, err = io.ReadFull(plc, buff)
		if err != nil {
			return nil, err
		}

		return nil, ErrorSelect(errorCode)
	}

	if retSize == 0 {
		return nil, nil
	}

	buff = make([]byte, retSize)

	_, err = io.ReadFull(plc, buff)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (plc *PlcConn) GetCPUInfo() (string, error) {
	cmd, err := plc.option.makeRequest(getCPUInfo())
	if err != nil {
		return "", err
	}

	_b, err := plc.SendCmd(cmd, 18)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(_b)), nil
}

func getCPUInfo() McMessage {
	return []byte{0x01, 0x01, 0x00, 0x00}
}
