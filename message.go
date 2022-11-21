package melsec

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var (
	// todo, manual melsec通信协议参考手册 P66
	SM_Component McMessage = []byte{0x91}
	SD_Component McMessage = []byte{0xA9}
	X_Component  McMessage = []byte{0x9C}
	Y_Component  McMessage = []byte{0x9D}
	M_Component  McMessage = []byte{0x90}
	L_Component  McMessage = []byte{0x92}
	F_Component  McMessage = []byte{0x93}
	V_Component  McMessage = []byte{0x94}
	B_Component  McMessage = []byte{0xA0}
	TN_Component McMessage = []byte{0xC2}
	D_Component  McMessage = []byte{0xA8}
	W_Component  McMessage = []byte{0xB4}
	TS_Component McMessage = []byte{0xC1}
	TC_Component McMessage = []byte{0xC0}
	R_Component  McMessage = []byte{0xAF}

	Base_10 int = 10
	Base_16 int = 16
)

type plcOptions struct {
	netCode               []byte
	plcCode               []byte
	targetModuleIoNo      []byte
	targetModuleStationNo []byte
	duration              []byte
}

func (plc plcOptions) generateMessage(device string, count int, values []byte) (McMessage, error) {
	isRead := len(values) == 0
	errorTitle := "generateMessage."

	// fixedpart + datapart
	fix := plc.getFixedPart()

	// datapart = datalength + command subcommand + device + count
	timer, err := plc.getCPUTimer()
	if err != nil {
		return nil, fmt.Errorf("get timer error: %s %s", errorTitle, err)
	}

	command := getSubCommand(isRead)

	request, err := generateMessage(isRead, device, count, values)
	if err != nil {
		return nil, fmt.Errorf("get request error: %s, %s", errorTitle, err)
	}

	dataBuff := bytes.Buffer{}
	dataBuff.Write(timer)
	dataBuff.Write(command)
	dataBuff.Write(request)

	// 请求数据长度为2字节
	requestlen, err := encodeUint(uint64(dataBuff.Len()), 2)
	if err != nil {
		return nil, fmt.Errorf("get request len error: %s %s", errorTitle, err)
	}

	total := bytes.Buffer{}
	total.Write(fix)
	total.Write(requestlen)

	dataBuff.WriteTo(&total)

	return total.Bytes(), nil
}

func (plc plcOptions) generateMessageMulti(device []string, count []int, values [][]byte) (McMessage, error) {
	isRead := len(values) == 0
	errorTitle := "generateMessageMulti"

	// fixedpart + datapart
	fix := plc.getFixedPart()

	// datapart = datalength + command subcommand + device + count
	timer, err := plc.getCPUTimer()
	if err != nil {
		return nil, fmt.Errorf("get timer error: %s, %s", errorTitle, err)
	}

	command := getSubCommandMulti(isRead)

	request, err := generateMessageMulti(isRead, device, count, values)
	if err != nil {
		return nil, fmt.Errorf("get request error: %s, %s", errorTitle, err)
	}

	dataBuff := bytes.Buffer{}
	dataBuff.Write(timer)
	dataBuff.Write(command)
	dataBuff.Write(request)

	// 请求数据长度为2字节
	requestlen, err := encodeUint(uint64(dataBuff.Len()), 2)
	if err != nil {
		return nil, fmt.Errorf("get request len error: %s, %s", errorTitle, err)
	}

	total := bytes.Buffer{}
	total.Write(fix)
	total.Write(requestlen)

	dataBuff.WriteTo(&total)

	return total.Bytes(), nil
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

type plcOption func(*plcOptions) error

func newplcOption(ops []plcOption) *plcOptions {
	opt := &plcOptions{
		netCode:               getLocalNetCode(),
		plcCode:               getPlcCode(),
		targetModuleIoNo:      getLocalTargetModuleIoNo(),
		targetModuleStationNo: getLocalTargetModuleStationNo(),
		duration:              getCPUTimer(),
	}

	for _, o := range ops {
		if o != nil {
			o(opt)
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

func SetNetCode(netcode interface{}) plcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, netcode.(uint)); err != nil {
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

func SetPLCCode(plcCode interface{}) plcOption {
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

func SetModuleIoNo(ioNo interface{}) plcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, ioNo.(uint16)); err != nil {
			return err
		}

		copy(opt.targetModuleIoNo, buff.Bytes())

		return nil
	}
}

func SetCPUTimer(t interface{}) plcOption {
	return func(opt *plcOptions) error {
		buff := bytes.Buffer{}

		if err := binary.Write(&buff, binary.LittleEndian, t.(uint16)); err != nil {
			return err
		}

		copy(opt.duration, (buff.Bytes()))

		return nil
	}
}

func getCPUTimer() McMessage {
	return []byte{0x0a, 0x00}
}

// getLocalTargetModuleStationNo, 返回请求目标站模块站编号.
func getLocalTargetModuleStationNo() McMessage {
	return []byte{0x00}
}

func SetModuleStationNo(stationNo interface{}) plcOption {
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

func (plc plcOptions) getCPUTimer() (McMessage, error) {
	return plc.duration, nil
}

func getSubCommand(isRead bool) McMessage {
	if isRead {
		return CommandMultiRead_Word_Binary
	}

	return CommandMultiWrite_Word_Binary
}

func getSubCommandMulti(isRead bool) McMessage {
	if isRead {
		return CommandMultiBlockRead_Binary
	}

	return CommandMultiBlockWrite_Binary
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
func generateMessage(isRead bool, device string, count int, values []byte) ([]byte, error) {
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

	if isRead {
		return b.Bytes(), nil
	} else {
		b.Write(values)
		return b.Bytes(), nil
	}
}

// softComponent + count + data.
func generateMessageMulti(isRead bool, device []string, count []int, values [][]byte) ([]byte, error) {
	re := make([]byte, 0)

	var wordCount, bitCount int8

	for i := 0; i < len(device); i++ {
		_compoType, _ := splitComponetName(device[i])
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

		if isRead {
			re = append(re, b.Bytes()...)
		} else {
			b.Write(values[i])
			re = append(re, b.Bytes()...)
		}
	}

	re = append([]byte{byte(wordCount), byte(bitCount)}, re...)

	return re, nil
}
