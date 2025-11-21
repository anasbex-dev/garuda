package utils

import (
    "encoding/binary"
    "math"
)

type BinaryStream struct {
    buffer []byte
    offset int
}

func NewBinaryStream(data []byte) *BinaryStream {
    return &BinaryStream{
        buffer: data,
        offset: 0,
    }
}

func (bs *BinaryStream) ReadByte() byte {
    if bs.offset >= len(bs.buffer) {
        return 0
    }
    value := bs.buffer[bs.offset]
    bs.offset++
    return value
}

func (bs *BinaryStream) ReadBytes(length int) []byte {
    if bs.offset+length > len(bs.buffer) {
        return nil
    }
    value := bs.buffer[bs.offset : bs.offset+length]
    bs.offset += length
    return value
}

func (bs *BinaryStream) ReadUint16() uint16 {
    data := bs.ReadBytes(2)
    if data == nil {
        return 0
    }
    return binary.BigEndian.Uint16(data)
}

func (bs *BinaryStream) ReadUint32() uint32 {
    data := bs.ReadBytes(4)
    if data == nil {
        return 0
    }
    return binary.BigEndian.Uint32(data)
}

func (bs *BinaryStream) ReadUint64() uint64 {
    data := bs.ReadBytes(8)
    if data == nil {
        return 0
    }
    return binary.BigEndian.Uint64(data)
}

func (bs *BinaryStream) ReadFloat32() float32 {
    data := bs.ReadBytes(4)
    if data == nil {
        return 0
    }
    return math.Float32frombits(binary.BigEndian.Uint32(data))
}

func (bs *BinaryStream) ReadString() string {
    length := bs.ReadUint16()
    data := bs.ReadBytes(int(length))
    if data == nil {
        return ""
    }
    return string(data)
}

func (bs *BinaryStream) WriteByte(value byte) {
    bs.buffer = append(bs.buffer, value)
}

func (bs *BinaryStream) WriteBytes(data []byte) {
    bs.buffer = append(bs.buffer, data...)
}

func (bs *BinaryStream) WriteUint16(value uint16) {
    data := make([]byte, 2)
    binary.BigEndian.PutUint16(data, value)
    bs.WriteBytes(data)
}

func (bs *BinaryStream) WriteUint32(value uint32) {
    data := make([]byte, 4)
    binary.BigEndian.PutUint32(data, value)
    bs.WriteBytes(data)
}

func (bs *BinaryStream) WriteString(value string) {
    bs.WriteUint16(uint16(len(value)))
    bs.WriteBytes([]byte(value))
}

func (bs *BinaryStream) Bytes() []byte {
    return bs.buffer
}

func (bs *BinaryStream) Reset() {
    bs.buffer = []byte{}
    bs.offset = 0
}