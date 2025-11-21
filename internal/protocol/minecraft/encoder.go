package minecraft

import (
    "encoding/binary"
    "garuda/pkg/utils"
)

type Encoder struct {
    stream *utils.BinaryStream
}

func NewEncoder() *Encoder {
    return &Encoder{
        stream: utils.NewBinaryStream(nil),
    }
}

func (e *Encoder) WriteVarInt(value int32) {
    for {
        if (value & ^0x7F) == 0 {
            e.stream.WriteByte(byte(value))
            return
        }
        e.stream.WriteByte(byte((value & 0x7F) | 0x80))
        value >>= 7
    }
}

func (e *Encoder) WriteString(value string) {
    e.WriteVarInt(int32(len(value)))
    e.stream.WriteBytes([]byte(value))
}

func (e *Encoder) WriteULong(value uint64) {
    data := make([]byte, 8)
    binary.BigEndian.PutUint64(data, value)
    e.stream.WriteBytes(data)
}

func (e *Encoder) WriteLong(value int64) {
    e.WriteULong(uint64(value))
}

func (e *Encoder) WriteUInt32(value uint32) {
    data := make([]byte, 4)
    binary.BigEndian.PutUint32(data, value)
    e.stream.WriteBytes(data)
}

func (e *Encoder) WriteInt32(value int32) {
    e.WriteUInt32(uint32(value))
}

func (e *Encoder) WriteShort(value int16) {
    data := make([]byte, 2)
    binary.BigEndian.PutUint16(data, uint16(value))
    e.stream.WriteBytes(data)
}

func (e *Encoder) WriteUShort(value uint16) {
    data := make([]byte, 2)
    binary.BigEndian.PutUint16(data, value)
    e.stream.WriteBytes(data)
}

func (e *Encoder) WriteFloat32(value float32) {
    data := make([]byte, 4)
    binary.BigEndian.PutUint32(data, math.Float32bits(value))
    e.stream.WriteBytes(data)
}

func (e *Encoder) WriteFloat64(value float64) {
    data := make([]byte, 8)
    binary.BigEndian.PutUint64(data, math.Float64bits(value))
    e.stream.WriteBytes(data)
}

func (e *Encoder) WriteBool(value bool) {
    if value {
        e.stream.WriteByte(1)
    } else {
        e.stream.WriteByte(0)
    }
}

func (e *Encoder) WriteByteArray(data []byte) {
    e.WriteVarInt(int32(len(data)))
    e.stream.WriteBytes(data)
}

func (e *Encoder) Bytes() []byte {
    return e.stream.Bytes()
}