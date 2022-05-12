package melsec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
)

var (
	// SubCommandMultiRead_Bit_Binary  McMessage = []byte{0x01, 0x04, 0x01, 0x00}

	CommandMultiRead_Word_Binary McMessage = []byte{0x01, 0x04, 0x00, 0x00}

	// SubCommandMultiWrite_Bit_Binary  McMessage = []byte{0x01, 0x14, 0x01, 0x00}

	CommandMultiWrite_Word_Binary McMessage = []byte{0x01, 0x14, 0x00, 0x00}

	// CommandRandomRead_Word_Binary McMessage = []byte{0x01, 0x04, 0x00, 0x00}

	CommandMultiBlockRead_Binary  McMessage = []byte{0x06, 0x04, 0x00, 0x00}
	CommandMultiBlockWrite_Binary McMessage = []byte{0x06, 0x14, 0x00, 0x00}
)

const (
	FirstResponseLength     = 11
	ResponseErrorCodeIndex  = 9
	ResponseErrorCodeLength = 2
)

type DeviceType uint8

const (
	EventDevice DeviceType = iota
	AckDevice
	StatusEventDevice
	AlarmEventDevice
	ReceiveEventDevice
	RemoveEventDevice
	SendEventDevice
)

type PlcConn struct {
	// conn   net.Conn
	net.Conn
	option *plcOptions
}

type Device struct {
	lock        sync.Mutex
	name        string
	count       int
	value       []byte
	mValue      []byte
	readMessage McMessage
	Error       error
	conn        *PlcConn
	changed     bool
}

func NewDevice(name string, count int, plc *PlcConn) *Device {
	return &Device{
		name:        name,
		count:       count,
		value:       make([]byte, 0, count*2),
		readMessage: make([]byte, 0),
		Error:       nil,
		conn:        plc,
	}
}

func (dev *Device) Count() int {
	return dev.count
}

func (dev *Device) Changed() bool {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	return dev.changed
}

func (dev *Device) Write() error {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if dev.mValue == nil {
		return nil
	}

	message, err := dev.conn.option.generateMessage(dev.name, dev.count, dev.mValue)
	if err != nil {
		return err
	}

	if _, err := dev.conn.Write(message); err != nil {
		return err
	}

	buff := make([]byte, FirstResponseLength)

	n, err := dev.conn.Read(buff)
	if err != nil && err != io.EOF {
		return err
	}

	if dev.mValue != nil {
		copy(dev.value, dev.mValue)

		dev.changed = true

		dev.mValue = nil
	}

	// 一次读取数据长度异常
	if n != FirstResponseLength {
		return fmt.Errorf("%s", "wrong response length")
	}

	// 返回错误代码
	if errorCode := buff[ResponseErrorCodeIndex : ResponseErrorCodeIndex+2]; !reflect.DeepEqual(errorCode, []byte{0x00, 0x00}) {
		// fmt.Printf("%x\n", errorCode)
		buff = make([]byte, ResponseErrorCodeLength)

		n, err = dev.conn.Read(buff)
		if n != ResponseErrorCodeLength {
			return errors.New("wrong response error length")
		}

		if err != nil && err != io.EOF {
			return err
		}

		return fmt.Errorf("errorCode: %x, error Info：%x", errorCode, buff)
	}

	return nil
}

func (dev *Device) Read() error {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if len(dev.readMessage) == 0 {
		message, err := dev.conn.option.generateMessage(dev.name, dev.count, nil)
		if err != nil {
			return err
		}

		dev.readMessage = message
	}

	if _, err := dev.conn.Write(dev.readMessage); err != nil {
		return err
	}

	buff := make([]byte, FirstResponseLength)

	n, err := dev.conn.Read(buff)
	if err != nil {
		if err == io.EOF {
			return errors.New("connection closed wrongly")
		}
		return err
	}

	// 一次读取数据长度异常
	if n != FirstResponseLength {
		return fmt.Errorf("%s", "wrong response length")
	}

	// 返回错误代码
	if errorCode := buff[ResponseErrorCodeIndex : ResponseErrorCodeIndex+2]; !reflect.DeepEqual(errorCode, []byte{0x00, 0x00}) {
		buff = make([]byte, ResponseErrorCodeLength)

		n, err = dev.conn.Read(buff)
		if n != ResponseErrorCodeLength {
			return fmt.Errorf("%s:", "wrong response error length")
		}

		if err != nil && err != io.EOF {
			return err
		}

		return fmt.Errorf("errorCode: %x, error Info：%x", errorCode, buff)
	}

	buff = make([]byte, dev.count*2)

	n, err = dev.conn.Read(buff)
	if n != dev.count*2 {
		return fmt.Errorf("%s", "wrong response data length")
	}

	if err != nil && err != io.EOF {
		return err
	}

	if reflect.DeepEqual(dev.value, buff) {
		return nil
	}

	dev.value = buff
	dev.changed = true

	return nil
}

func (dev *Device) GetValue() []byte {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	dev.changed = false

	return dev.value
}

func (dev *Device) SetValue(val []byte) {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	buffer := bytes.NewBuffer(nil)
	binary.Write(buffer, binary.LittleEndian, val)
	for buffer.Len() < dev.count*2 {
		binary.Write(buffer, binary.LittleEndian, uint16(0))
	}

	dev.mValue = buffer.Bytes()
	dev.changed = false
}

func (dev *Device) Name() string {
	return dev.name
}

func NewConn(addr, port string, errChan chan error, ops ...plcOption) (*PlcConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":"+port)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	conn.SetKeepAlive(true)

	return &PlcConn{
		conn,
		newplcOption(ops),
	}, nil
}

type MultiDevice struct {
	lock        sync.Mutex
	name        []string
	count       []int
	value       [][]byte
	mValue      [][]byte
	readMessage McMessage
	Error       error
	conn        *PlcConn
	changed     bool
}

func (dev *MultiDevice) Count() []int {
	return dev.count
}

func (dev *MultiDevice) Changed() bool {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	return dev.changed
}

func (dev *MultiDevice) Write() error {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if dev.mValue == nil {
		return nil
	}

	message, err := dev.conn.option.generateMessageMulti(dev.name, dev.count, dev.mValue)
	if err != nil {
		return err
	}

	if _, err := dev.conn.Write(message); err != nil {
		return err
	}

	buff := make([]byte, FirstResponseLength)

	n, err := dev.conn.Read(buff)
	if err != nil && err != io.EOF {
		return err
	}

	if dev.mValue != nil {
		copy(dev.value, dev.mValue)

		dev.changed = true

		dev.mValue = nil
	}

	// 一次读取数据长度异常
	if n != FirstResponseLength {
		return fmt.Errorf("%s", "wrong response length")
	}

	// 返回错误代码
	if errorCode := buff[ResponseErrorCodeIndex : ResponseErrorCodeIndex+2]; !reflect.DeepEqual(errorCode, []byte{0x00, 0x00}) {
		buff = make([]byte, ResponseErrorCodeLength)

		n, err = dev.conn.Read(buff)
		if n != ResponseErrorCodeLength {
			return errors.New("wrong response error length")
		}

		if err != nil && err != io.EOF {
			return err
		}

		return fmt.Errorf("errorCode: %x, error Info：%x", errorCode, buff)
	}

	return nil
}

func (dev *MultiDevice) Read() error {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if len(dev.readMessage) == 0 {
		message, err := dev.conn.option.generateMessageMulti(dev.name, dev.count, nil)
		if err != nil {
			return err
		}

		dev.readMessage = message
	}

	if _, err := dev.conn.Write(dev.readMessage); err != nil {
		return err
	}

	buff := make([]byte, FirstResponseLength)

	n, err := dev.conn.Read(buff)
	if err != nil {
		if err == io.EOF {
			return errors.New("connection closed wrongly")
		}
		return err
	}

	// 一次读取数据长度异常
	if n != FirstResponseLength {
		return fmt.Errorf("%s", "wrong response length")
	}

	// 返回错误代码
	if errorCode := buff[ResponseErrorCodeIndex : ResponseErrorCodeIndex+2]; !reflect.DeepEqual(errorCode, []byte{0x00, 0x00}) {
		buff = make([]byte, ResponseErrorCodeLength)

		n, err = dev.conn.Read(buff)
		if n != ResponseErrorCodeLength {
			return fmt.Errorf("%s:", "wrong response error length")
		}

		if err != nil && err != io.EOF {
			return err
		}

		return fmt.Errorf("errorCode: %x, error Info：%x", errorCode, buff)
	}

	totalCount := dev.totalCount() * 2

	buff = make([]byte, totalCount)

	n, err = dev.conn.Read(buff)
	if n != totalCount {
		fmt.Printf("read: % x", buff)

		return fmt.Errorf("%s %d %d ", "wrong response data length", n, totalCount)
	}

	if err != nil && err != io.EOF {
		return err
	}

	if reflect.DeepEqual(dev.value, buff) {
		return nil
	}

	for i := 0; i < len(dev.count); i++ {
		dev.value[i] = buff[:dev.count[i]]
		buff = buff[dev.count[i]:]
	}

	dev.changed = true

	return nil
}

func (dev *MultiDevice) GetValue() [][]byte {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	dev.changed = false

	return dev.value
}

func (dev *MultiDevice) totalCount() int {
	totalCount := 0
	for i := range dev.count {
		totalCount += dev.count[i]
	}

	return totalCount
}

func (dev *MultiDevice) SetValue(val [][]byte) {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	dev.mValue = val
	dev.changed = false
}

func (dev *MultiDevice) Name() []string {
	return dev.name
}

func (dev *MultiDevice) AddBlock(name string, count int) {
	dev.name = append(dev.name, name)
	dev.count = append(dev.count, count)
	dev.value = append(dev.value, make([]byte, 0))
	dev.mValue = append(dev.mValue, make([]byte, 0))
}

func NewMultiDevice(conn *PlcConn) *MultiDevice {
	return &MultiDevice{
		lock:        sync.Mutex{},
		name:        make([]string, 0),
		count:       make([]int, 0),
		value:       make([][]byte, 0),
		mValue:      make([][]byte, 0),
		readMessage: make([]byte, 0),
		Error:       nil,
		conn:        conn,
		changed:     false,
	}
}
