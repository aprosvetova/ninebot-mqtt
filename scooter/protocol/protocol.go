package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const ReadRequestCommand RequestCommand = 0x01

func GetBattery() []byte {
	return CreateRequest(ReadRequestCommand, 0x22, 0x02)
}

func CreateRequest(command RequestCommand, param byte, payload... byte) []byte {
	cmd := []byte{0x5A, 0xA5, 0x00, 0x3E, 0x20, byte(command), param}
	cmd = append(cmd, payload...)
	cmd[2] = byte(len(payload))
	cmd = append(cmd, getChecksum(cmd[2:])...)
	return cmd
}

func ParseResponse(raw []byte) (*Response, error) {
	if len(raw) < 9 {
		return nil, errors.New("raw is too short")
	}
	if raw[0] != 0x5A || raw[1] != 0xA5 || raw[3] != 0x20 || raw[4] != 0x3E {
		return nil, errors.New("not a Ninebot ES raw")
	}
	if raw[2] != uint8(len(raw) - 9) {
		return nil, errors.New("wrong payload length byte")
	}
	if bytes.Compare(raw[len(raw)-2:], getChecksum(raw[2:len(raw)-2])) != 0 {
		return nil, errors.New("wrong checksum")
	}
	response := &Response{
		Command: raw[5],
		Parameter: raw[6],
		Payload: raw[7:len(raw)-2],
	}
	return response, nil
}

func getChecksum(part []byte) []byte {
	chkSum := 0xFFFF
	for _, b := range part {
		chkSum -= int(b)
	}
	bChkSum := make([]byte, 2)
	binary.LittleEndian.PutUint16(bChkSum, uint16(chkSum))
	return bChkSum
}