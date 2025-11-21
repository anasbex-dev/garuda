package minecraft

import (
    "encoding/binary"
    "garuda/pkg/utils"
    "math"
)

type Decoder struct {
    stream *utils.BinaryStream
}

func NewDecoder(data []byte) *Decoder {
    return &Decoder{
        stream: utils.NewBinaryStream(data),
    }
}

func (d *Decoder) ReadVarInt() int32 {
    var value int32
    var position int
    var currentByte byte

    for {
        currentByte = d.stream.ReadByte()
        value |= int32(currentByte&0x7F) << position

        if (currentByte & 0x80) == 0 {
            break
        }

        position += 7
        if position >= 32 {
            return 0 // Error
        }
    }

    return value
}

func (d *Decoder) ReadString() string {
    length := d.ReadVarInt()
    if length <= 0 {
        return ""
    }
    data := d.stream.ReadBytes(int(length))
    return string(data)
}

func (d *Decoder) ReadULong() uint64 {
    data := d.stream.ReadBytes(8)
    if len(data) < 8 {
        return 0
    }
    return binary.BigEndian.Uint64(data)
}

func (d *Decoder) ReadLong() int64 {
    return int64(d.ReadULong())
}

func (d *Decoder) ReadUInt32() uint32 {
    data := d.stream.ReadBytes(4)
    if len(data) < 4 {
        return 0
    }
    return binary.BigEndian.Uint32(data)
}

func (d *Decoder) ReadInt32() int32 {
    return int32(d.ReadUInt32())
}

func (d *Decoder) ReadShort() int16 {
    data := d.stream.ReadBytes(2)
    if len(data) < 2 {
        return 0
    }
    return int16(binary.BigEndian.Uint16(data))
}

func (d *Decoder) ReadUShort() uint16 {
    data := d.stream.ReadBytes(2)
    if len(data) < 2 {
        return 0
    }
    return binary.BigEndian.Uint16(data)
}

func (d *Decoder) ReadFloat32() float32 {
    data := d.stream.ReadBytes(4)
    if len(data) < 4 {
        return 0
    }
    return math.Float32frombits(binary.BigEndian.Uint32(data))
}

func (d *Decoder) ReadFloat64() float64 {
    data := d.stream.ReadBytes(8)
    if len(data) < 8 {
        return 0
    }
    return math.Float64frombits(binary.BigEndian.Uint64(data))
}

func (d *Decoder) ReadBool() bool {
    return d.stream.ReadByte() == 1
}

func (d *Decoder) ReadByteArray() []byte {
    length := d.ReadVarInt()
    if length <= 0 {
        return nil
    }
    return d.stream.ReadBytes(int(length))
}

func (d *Decoder) Remaining() int {
    return len(d.stream.Bytes()) - d.stream.Offset
}