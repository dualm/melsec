package melsec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"sync"
)

type DeviceType uint8

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

func NewDevice(name string, count int, plc *PlcConn) (*Device, error) {
	if name == "" {
		return nil, errors.New("empty address")
	}
	if count == 0 {
		return nil, errors.New("count is 0")
	}
	if plc == nil {
		return nil, errors.New("nil plc connection")
	}
	return &Device{
		name:        name,
		count:       count,
		value:       make([]byte, 0, count*2),
		readMessage: make([]byte, 0),
		Error:       nil,
		conn:        plc,
	}, nil
}

func (dev *Device) Count() int {
	return dev.count
}

func (dev *Device) Changed() bool {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	return dev.changed
}

// Write 执行写入操作, 写入内容为最近一次SetValue时传入的值.
func (dev *Device) Write() error {

	if dev.mValue == nil {
		return nil
	}

	message, err := dev.conn.option.generateMessage(dev.name, dev.count, dev.mValue)
	if err != nil {
		return err
	}

	_, err = dev.conn.SendCmd(message, 0)
	if err != nil {
		return err
	}

	// 更新数据
	dev.lock.Lock()
	copy(dev.value, dev.mValue)
	dev.changed = true
	dev.mValue = nil
	dev.lock.Unlock()

	return nil
}

func (dev *Device) Read() error {
	if len(dev.readMessage) == 0 {
		message, err := dev.conn.option.generateMessage(dev.name, dev.count, nil)
		if err != nil {
			return err
		}

		dev.readMessage = message
	}

	buff, err := dev.conn.SendCmd(dev.readMessage, dev.count*2)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(dev.value, buff) {
		return nil
	}

	dev.lock.Lock()
	dev.value = buff
	dev.changed = true
	dev.lock.Unlock()

	return nil
}

func (dev *Device) GetValue() []byte {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	dev.changed = false

	return dev.value
}

func (dev *Device) SetValue(val []byte) {
	buffer := bytes.NewBuffer(nil)
	_ = binary.Write(buffer, binary.LittleEndian, val)

	for buffer.Len() < dev.count*2 {
		_ = binary.Write(buffer, binary.LittleEndian, uint16(0))
	}

	dev.lock.Lock()
	dev.mValue = buffer.Bytes()
	dev.changed = false
	dev.lock.Unlock()
}

func (dev *Device) Name() string {
	return dev.name
}
