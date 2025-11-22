package minecraft

import (
	"encoding/binary"
	"errors"
	"garuda/pkg/utils"
)

const (
	MAX_PACKET_SIZE = 1024 * 1024 // 1MB
)

type PacketDecoder struct {
	compressionEnabled bool
	compressionThreshold int
}

func NewPacketDecoder() *PacketDecoder {
	return &PacketDecoder{
		compressionEnabled: false,
		compressionThreshold: 512,
	}
}

func (d *PacketDecoder) DecodePacket(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, errors.New("empty packet data")
	}

	// Minecraft Bedrock uses length-prefixed packets
	offset := 0
	length, bytesRead := binary.Uvarint(data[offset:])
	if bytesRead <= 0 {
		return nil, errors.New("invalid packet length")
	}
	offset += bytesRead

	if length > MAX_PACKET_SIZE {
		return nil, errors.New("packet too large")
	}

	if int(length) > len(data)-offset {
		return nil, errors.New("incomplete packet")
	}

	packetData := data[offset : offset+int(length)]
	return packetData, nil
}

func (d *PacketDecoder) EncodePacket(data []byte) ([]byte, error) {
	lengthBuf := make([]byte, binary.MaxVarintLen32)
	bytesWritten := binary.PutUvarint(lengthBuf, uint64(len(data)))
	
	result := make([]byte, bytesWritten+len(data))
	copy(result, lengthBuf[:bytesWritten])
	copy(result[bytesWritten:], data)
	
	return result, nil
}