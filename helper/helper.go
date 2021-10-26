package helper

import (
	"encoding/binary"
	"os"
)

func ReadByte(file *os.File) (byte, error) {
	var buffer [1]byte
	_, err := file.Read(buffer[:])
	if err != nil {
		return 0, err
	}
	return buffer[0], nil
}

func ReadUint16(file *os.File) (uint16, error) {
	var buffer [2]byte
	_, err := file.Read(buffer[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buffer[:]), nil
}

func ReadUint32(file *os.File) (uint32, error) {
	var buffer [4]byte
	_, err := file.Read(buffer[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buffer[:]), nil
}

func ReadUint64(file *os.File) (uint64, error) {
	var buffer [8]byte
	_, err := file.Read(buffer[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buffer[:]), nil
}
