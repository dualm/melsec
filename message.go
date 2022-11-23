package melsec

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	// Manual melsec通信协议参考手册 P66

	SMComponent McMessage = []byte{0x91}
	SDComponent McMessage = []byte{0xA9}
	XComponent  McMessage = []byte{0x9C}
	YComponent  McMessage = []byte{0x9D}
	MComponent  McMessage = []byte{0x90}
	LComponent  McMessage = []byte{0x92}
	FComponent  McMessage = []byte{0x93}
	VComponent  McMessage = []byte{0x94}
	BComponent  McMessage = []byte{0xA0}
	TNComponent McMessage = []byte{0xC2}
	DComponent  McMessage = []byte{0xA8}
	WComponent  McMessage = []byte{0xB4}
	TSComponent McMessage = []byte{0xC1}
	TCComponent McMessage = []byte{0xC0}
	RComponent  McMessage = []byte{0xAF}

	Base10 int = 10
	Base16 int = 16
)

type plcOptions struct {
	netCode               []byte
	plcCode               []byte
	targetModuleIoNo      []byte
	targetModuleStationNo []byte
	duration              []byte
}

func (plc plcOptions) makeRequest(cmd McMessage) (McMessage, error) {
	fix := plc.getFixedPart()

	timer := plc.getCPUTimer()

	buf := bytes.NewBuffer(nil)

	buf.Write(timer)
	buf.Write(cmd)

	requestLen, err := encodeUint(uint64(buf.Len()), 2)
	if err != nil {
		return nil, fmt.Errorf("get request len error: %w", err)
	}

	total := bytes.NewBuffer(nil)

	_, err = total.Write(fix)
	if err != nil {
		return nil, fmt.Errorf("write fix part to request error, %w", err)
	}

	_, err = total.Write(requestLen)
	if err != nil {
		return nil, fmt.Errorf("write len to request error, %w", err)
	}

	_, err = buf.WriteTo(total)
	if err != nil {
		return nil, fmt.Errorf("write cmd to request error, %w", err)
	}

	return total.Bytes(), nil
}

func (plc plcOptions) generateMessage(device string, count int, values []byte) (McMessage, error) {
	command := getSubOperation(len(values) == 0)

	request, err := generateCmd(device, count, values)
	if err != nil {
		return nil, fmt.Errorf("get request error: %w", err)
	}

	dataBuff := bytes.Buffer{}
	dataBuff.Write(command)
	dataBuff.Write(request)

	return plc.makeRequest(dataBuff.Bytes())
}

func (plc plcOptions) generateMessageMulti(device []string, count []int, values [][]byte) (McMessage, error) {
	command := getSubCommandMulti(len(values) == 0)

	request, err := generateCmdMulti(device, count, values)
	if err != nil {
		return nil, fmt.Errorf("get request error:  %s", err)
	}

	dataBuff := bytes.Buffer{}
	dataBuff.Write(command)
	dataBuff.Write(request)

	return plc.makeRequest(dataBuff.Bytes())
}

func (plc plcOptions) getNetCode() McMessage {
	return plc.netCode
}

func (plc plcOptions) getPlcCode() McMessage {
	return plc.plcCode
}

func (plc plcOptions) getTargetModuleIoNo() McMessage {
	return plc.targetModuleIoNo
}

func (plc plcOptions) getTargetModuleStationNo() McMessage {
	return plc.targetModuleStationNo
}

type PlcOption func(*plcOptions) error

func newPlcOption(ops []PlcOption) *plcOptions {
	opt := &plcOptions{
		netCode:               getLocalNetCode(),
		plcCode:               getPlcCode(),
		targetModuleIoNo:      getLocalTargetModuleIoNo(),
		targetModuleStationNo: getLocalTargetModuleStationNo(),
		duration:              getCPUTimer(),
	}

	for _, o := range ops {
		if o != nil {
			_ = o(opt)
		}
	}

	return opt
}

// getSubtitle，返回副帧头.
// 4E: []byte{0x54, 0x00}
// 3E: []byte{0x50, 0x00}
func getSubtitle() McMessage {
	return []byte{0x50, 0x00}
}

// getLocalNetCode, 返回访问站网络号
func getLocalNetCode() McMessage {
	return []byte{0x00}
}

func SetNetCode(netCode interface{}) PlcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, netCode.(uint)); err != nil {
			return err
		}

		copy(opt.netCode, buff.Bytes())

		return nil
	}
}

// getPlcCode,返回访问站PLC编号.
func getPlcCode() McMessage {
	return []byte{0xFF}
}

func SetPLCCode(plcCode interface{}) PlcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}
		if err := binary.Write(&buff, binary.LittleEndian, plcCode.(uint)); err != nil {
			return err
		}

		copy(opt.plcCode, buff.Bytes())

		return nil
	}
}

// getTargetModuleIoNo, 返回请求目标模块IO编号.
func getLocalTargetModuleIoNo() McMessage {
	return []byte{0xFF, 0x03}
}

func SetModuleIoNo(ioNo interface{}) PlcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, ioNo.(uint16)); err != nil {
			return err
		}

		copy(opt.targetModuleIoNo, buff.Bytes())

		return nil
	}
}

func SetCPUTimer(t interface{}) PlcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, t.(uint16)); err != nil {
			return err
		}

		copy(opt.duration, buff.Bytes())

		return nil
	}
}

// getCPUTimer 初始化时生成默认的定时器时间
func getCPUTimer() McMessage {
	return []byte{0x01, 0x00}
}

// getLocalTargetModuleStationNo, 返回请求目标站模块站编号.
func getLocalTargetModuleStationNo() McMessage {
	return []byte{0x00}
}

func SetModuleStationNo(stationNo interface{}) PlcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, stationNo.(uint)); err != nil {
			return err
		}

		copy(opt.targetModuleStationNo, buff.Bytes())

		return nil
	}
}

type McMessage []byte

func (plc plcOptions) getFixedPart() McMessage {
	b := bytes.Buffer{}
	b.Write(getSubtitle())
	b.Write(plc.getNetCode())
	b.Write(plc.getPlcCode())
	b.Write(plc.getTargetModuleIoNo())
	b.Write(plc.getTargetModuleStationNo())

	return b.Bytes()
}

func (plc plcOptions) getCPUTimer() McMessage {
	return plc.duration
}

func getSubOperation(isRead bool) McMessage {
	if isRead {
		return CommandMultiReadWordBinary
	}

	return CommandMultiWriteWordBinary
}

func getSubCommandMulti(isRead bool) McMessage {
	if isRead {
		return CommandMultiBlockReadBinary
	}

	return CommandMultiBlockWriteBinary
}

func updateMcMessage(data interface{}) (McMessage, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// softComponent + count + data.
func generateCmd(device string, count int, values []byte) ([]byte, error) {
	sc, err := encodeSoftComponent(device)
	if err != nil {
		return nil, fmt.Errorf("generateMessage error: %s", err)
	}

	b := bytes.Buffer{}
	b.Write(sc)

	softComponentCount, err := encodeUint(uint64(count), 2)
	if err != nil {
		return nil, fmt.Errorf("generateMessage error: %s", err)
	}

	b.Write(softComponentCount)

	if len(values) != 0 {
		b.Write(values)
		return b.Bytes(), nil
	}

	return b.Bytes(), nil
}

// softComponent + count + data.
func generateCmdMulti(device []string, count []int, values [][]byte) ([]byte, error) {
	re := make([]byte, 0)

	var wordCount, bitCount int8

	for i := 0; i < len(device); i++ {
		_compoType, _ := splitComponentName(device[i])
		_bitSize, _wordSize := componentBitSize(_compoType)
		wordCount += _wordSize
		bitCount += _bitSize

		sc, err := encodeSoftComponent(device[i])
		if err != nil {
			return nil, fmt.Errorf("generateMessageMulti error: %w", err)
		}

		b := bytes.Buffer{}
		b.Write(sc)

		softComponentCount, err := encodeUint(uint64(count[i]), 2)
		if err != nil {
			return nil, fmt.Errorf("generateMessageMulti error: %w", err)
		}

		b.Write(softComponentCount)

		if len(values) != 0 {
			b.Write(values[i])
		}

		re = append(re, b.Bytes()...)
	}

	re = append([]byte{byte(wordCount), byte(bitCount)}, re...)

	return re, nil
}
