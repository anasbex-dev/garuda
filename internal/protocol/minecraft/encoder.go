package minecraft

import (
	"bytes"
	"compress/zlib"
	"garuda/pkg/utils"
	"io"
)

type PacketEncoder struct {
	compressionEnabled bool
	compressionThreshold int
}

func NewPacketEncoder() *PacketEncoder {
	return &PacketEncoder{
		compressionEnabled: false,
		compressionThreshold: 512,
	}
}

func (e *PacketEncoder) EnableCompression(threshold int) {
	e.compressionEnabled = true
	e.compressionThreshold = threshold
}

func (e *PacketEncoder) Encode(packet Packet) ([]byte, error) {
	packetData, err := packet.Encode()
	if err != nil {
		return nil, err
	}

	// Add packet ID
	fullData := make([]byte, 0, len(packetData)+1)
	fullData = append(fullData, packet.ID())
	fullData = append(fullData, packetData...)

	// Apply compression if enabled
	if e.compressionEnabled && len(fullData) >= e.compressionThreshold {
		return e.compressPacket(fullData)
	}

	return fullData, nil
}

func (e *PacketEncoder) compressPacket(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (e *PacketEncoder) DecompressPacket(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}