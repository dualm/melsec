package mock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

func Mock(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer func() {
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		if err := handleConn(conn); err != nil {
			return err
		}
	}
}

func handleConn(conn net.Conn) error {
	for {
		// 副帧头 + path
		_Header := make([]byte, 7)

		if n, err := conn.Read(_Header); err != nil {
			if err == io.EOF || n != 7 {
				return fmt.Errorf("消息头格式错误, 正确格式应该为副帧头(2字节)+网络号(1字节)+PLC号(1字节)+请求目标模块IO编号(2字节)+请求目标模块站号(1字节)")
			}

			return err
		}

		// Length
		_2Bytes := make([]byte, 2)

		if n, err := conn.Read(_2Bytes); err != nil {
			if err == io.EOF || n != 2 {
				return fmt.Errorf("消息长度错误, 正确格式应该为2个字节")
			}

			return err
		}

		_DataLength := binary.LittleEndian.Uint16(_2Bytes)

		_Data := make([]byte, _DataLength)

		if n, err := conn.Read(_Data); err != nil {
			if err == io.EOF || n != int(_DataLength) {
				return fmt.Errorf("消息指令区长度错误")
			}

			return err
		}

		_DataBuf := bytes.NewBuffer(_Data)

		log.Printf("data: % x", _DataBuf.Bytes())
		// timer
		_, _ = _DataBuf.Read(_2Bytes)
		// cmd
		_, _ = _DataBuf.Read(_2Bytes)

		log.Printf("指令：% x", _2Bytes)
		// cmd
		_, _ = _DataBuf.Read(_2Bytes)

		// 起始软元件编号
		_StartSoftNo := make([]byte, 3)
		_, _ = _DataBuf.Read(_StartSoftNo)
		_StartSoftNo = append(_StartSoftNo, make([]byte, 5)...)

		log.Printf("地址： %d ", binary.LittleEndian.Uint64(_StartSoftNo))
	}
}
