package melsec

import (
	"errors"
	"reflect"
	"sync"
)

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

	_, err = dev.conn.SendCmd(message, 0)
	if err != nil {
		return err
	}

	// 更新数据
	copy(dev.value, dev.mValue)
	dev.changed = true
	dev.mValue = nil

	return nil
}

func (dev *MultiDevice) getReadMessage() (McMessage, error) {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if len(dev.readMessage) == 0 {
		message, err := dev.conn.option.generateMessageMulti(dev.name, dev.count, nil)
		if err != nil {
			return nil, err
		}

		dev.readMessage = message
	}

	return dev.readMessage, nil
}

func (dev *MultiDevice) Read() error {
	msg, err := dev.getReadMessage()
	if err != nil {
		return err
	}

	buff, err := dev.conn.SendCmd(msg, dev.totalCount()*2)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(dev.value, buff) {
		return nil
	}

	dev.lock.Lock()
	defer dev.lock.Unlock()

	for i := 0; i < len(dev.count); i++ {
		dev.value[i] = buff[:dev.count[i]*2]
		buff = buff[dev.count[i]*2:]
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

func NewMultiDevice(conn *PlcConn) (*MultiDevice, error) {
	if conn == nil {
		return nil, errors.New("nil plc connection")
	}

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
	}, nil
}
