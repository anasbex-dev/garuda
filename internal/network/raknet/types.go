package raknet

import (
    "time"
)

// PacketReliability defines the reliability of packets
type PacketReliability int

const (
    ReliabilityUnreliable PacketReliability = iota
    ReliabilityUnreliableSequenced
    ReliabilityReliable
    ReliabilityReliableOrdered
    ReliabilityReliableSequenced
    ReliabilityUnreliableWithAckReceipt
    ReliabilityReliableWithAckReceipt
    ReliabilityReliableOrderedWithAckReceipt
)

// String returns the string representation of PacketReliability
func (r PacketReliability) String() string {
    switch r {
    case ReliabilityUnreliable:
        return "UNRELIABLE"
    case ReliabilityUnreliableSequenced:
        return "UNRELIABLE_SEQUENCED"
    case ReliabilityReliable:
        return "RELIABLE"
    case ReliabilityReliableOrdered:
        return "RELIABLE_ORDERED"
    case ReliabilityReliableSequenced:
        return "RELIABLE_SEQUENCED"
    case ReliabilityUnreliableWithAckReceipt:
        return "UNRELIABLE_WITH_ACK_RECEIPT"
    case ReliabilityReliableWithAckReceipt:
        return "RELIABLE_WITH_ACK_RECEIPT"
    case ReliabilityReliableOrderedWithAckReceipt:
        return "RELIABLE_ORDERED_WITH_ACK_RECEIPT"
    default:
        return "UNKNOWN"
    }
}

// EncapsulatedPacket represents a packet within a datagram
type EncapsulatedPacket struct {
    Reliability    PacketReliability
    HasSplit       bool
    MessageIndex   uint32
    SequenceIndex  uint32
    OrderIndex     uint32
    OrderChannel   byte
    SplitCount     uint32
    SplitID        uint16
    SplitIndex     uint32
    Data           []byte
    Length         uint16
    Timestamp      time.Time
}

// Datagram represents a RakNet datagram containing multiple packets
type Datagram struct {
    SequenceNumber uint32
    Packets        []*EncapsulatedPacket
    Timestamp      time.Time
}

// ACKRange represents a range of acknowledged sequences
type ACKRange struct {
    Start uint32
    End   uint32
}

// NACKRange represents a range of missing sequences
type NACKRange struct {
    Start uint32
    End   uint32
}

// ACKPacket represents an acknowledgement packet
type ACKPacket struct {
    Ranges []ACKRange
}

// NACKPacket represents a negative acknowledgement packet
type NACKPacket struct {
    Ranges []NACKRange
}

// ConnectionState represents the state of a RakNet connection
type ConnectionState int

const (
    StateUnconnected ConnectionState = iota
    StateRequestingConnection
    StateHandshaking
    StateConnected
    StateDisconnecting
    StateDisconnected
)

// String returns the string representation of ConnectionState
func (s ConnectionState) String() string {
    switch s {
    case StateUnconnected:
        return "UNCONNECTED"
    case StateRequestingConnection:
        return "REQUESTING_CONNECTION"
    case StateHandshaking:
        return "HANDSHAKING"
    case StateConnected:
        return "CONNECTED"
    case StateDisconnecting:
        return "DISCONNECTING"
    case StateDisconnected:
        return "DISCONNECTED"
    default:
        return "UNKNOWN"
    }
}

// PacketPriority defines the priority of packets
type PacketPriority int

const (
    PriorityImmediate PacketPriority = iota
    PriorityHigh
    PriorityMedium
    PriorityLow
)

// String returns the string representation of PacketPriority
func (p PacketPriority) String() string {
    switch p {
    case PriorityImmediate:
        return "IMMEDIATE"
    case PriorityHigh:
        return "HIGH"
    case PriorityMedium:
        return "MEDIUM"
    case PriorityLow:
        return "LOW"
    default:
        return "UNKNOWN"
    }
}

// PacketID represents RakNet packet identifiers
type PacketID byte

const (
    IDConnectedPing          PacketID = 0x00
    IDUnconnectedPing        PacketID = 0x01
    IDUnconnectedPingOpen    PacketID = 0x02
    IDConnectedPong          PacketID = 0x03
    IDDetectLostConnections  PacketID = 0x04
    IDOpenConnectionRequest1 PacketID = 0x05
    IDOpenConnectionReply1   PacketID = 0x06
    IDOpenConnectionRequest2 PacketID = 0x07
    IDOpenConnectionReply2   PacketID = 0x08
    IDConnectionRequest      PacketID = 0x09
    IDConnectionRequestAccepted PacketID = 0x10
    IDConnectionAttemptFailed   PacketID = 0x11
    IDAlreadyConnected       PacketID = 0x12
    IDNewIncomingConnection  PacketID = 0x13
    IDNoFreeIncomingConnections PacketID = 0x14
    IDDisconnectionNotification PacketID = 0x15
    IDConnectionLost         PacketID = 0x16
    IDConnectionBanned       PacketID = 0x17
    IDIncompatibleProtocol   PacketID = 0x18
    IDUnconnectedPong        PacketID = 0x1c
    IDAdvertiseSystem        PacketID = 0x1d
    IDUserPacketEnum         PacketID = 0x80
)

// String returns the string representation of PacketID
func (p PacketID) String() string {
    switch p {
    case IDConnectedPing:
        return "CONNECTED_PING"
    case IDUnconnectedPing:
        return "UNCONNECTED_PING"
    case IDUnconnectedPingOpen:
        return "UNCONNECTED_PING_OPEN"
    case IDConnectedPong:
        return "CONNECTED_PONG"
    case IDDetectLostConnections:
        return "DETECT_LOST_CONNECTIONS"
    case IDOpenConnectionRequest1:
        return "OPEN_CONNECTION_REQUEST_1"
    case IDOpenConnectionReply1:
        return "OPEN_CONNECTION_REPLY_1"
    case IDOpenConnectionRequest2:
        return "OPEN_CONNECTION_REQUEST_2"
    case IDOpenConnectionReply2:
        return "OPEN_CONNECTION_REPLY_2"
    case IDConnectionRequest:
        return "CONNECTION_REQUEST"
    case IDConnectionRequestAccepted:
        return "CONNECTION_REQUEST_ACCEPTED"
    case IDConnectionAttemptFailed:
        return "CONNECTION_ATTEMPT_FAILED"
    case IDAlreadyConnected:
        return "ALREADY_CONNECTED"
    case IDNewIncomingConnection:
        return "NEW_INCOMING_CONNECTION"
    case IDNoFreeIncomingConnections:
        return "NO_FREE_INCOMING_CONNECTIONS"
    case IDDisconnectionNotification:
        return "DISCONNECTION_NOTIFICATION"
    case IDConnectionLost:
        return "CONNECTION_LOST"
    case IDConnectionBanned:
        return "CONNECTION_BANNED"
    case IDIncompatibleProtocol:
        return "INCOMPATIBLE_PROTOCOL"
    case IDUnconnectedPong:
        return "UNCONNECTED_PONG"
    case IDAdvertiseSystem:
        return "ADVERTISE_SYSTEM"
    case IDUserPacketEnum:
        return "USER_PACKET_ENUM"
    default:
        return "UNKNOWN"
    }
}

// NetworkStatistics holds statistics for a connection
type NetworkStatistics struct {
    BytesSent            uint64
    BytesReceived        uint64
    PacketsSent          uint64
    PacketsReceived      uint64
    PacketsLost          uint64
    MessagesInSendBuffer uint32
    BytesInSendBuffer    uint32
    MessagesInResendBuffer uint32
    BytesInResendBuffer  uint32
    PacketLoss           float32
    LastUpdate           time.Time
}

// ConnectionMetrics holds connection performance metrics
type ConnectionMetrics struct {
    Ping                 time.Duration
    Jitter               time.Duration
    PacketLoss           float32
    Throttle             uint32
    MTU                  uint32
    BandwidthLimit       uint32
}

// SplitPacketInfo holds information about split packets
type SplitPacketInfo struct {
    SplitID    uint16
    SplitCount uint16
    SplitIndex uint16
    Timestamp  time.Time
}

// MessageOrderingInfo holds information for ordered messages
type MessageOrderingInfo struct {
    OrderChannel byte
    OrderIndex   uint32
}

// ReliabilityLayerConfig holds configuration for the reliability layer
type ReliabilityLayerConfig struct {
    MaxRetries           uint32
    RetryDelay           time.Duration
    AckDelay             time.Duration
    MaxResendBufferSize  uint32
    MaxSendBufferSize    uint32
    SplitPacketChannel   byte
}

// DefaultReliabilityLayerConfig returns the default configuration
func DefaultReliabilityLayerConfig() *ReliabilityLayerConfig {
    return &ReliabilityLayerConfig{
        MaxRetries:           5,
        RetryDelay:           100 * time.Millisecond,
        AckDelay:             10 * time.Millisecond,
        MaxResendBufferSize:  1024 * 1024, // 1MB
        MaxSendBufferSize:    2 * 1024 * 1024, // 2MB
        SplitPacketChannel:   0,
    }
}

// PacketQueueItem represents an item in the packet queue
type PacketQueueItem struct {
    Packet      *EncapsulatedPacket
    Priority    PacketPriority
    Timestamp   time.Time
    Retries     uint32
}

// ConnectionRequest contains connection request data
type ConnectionRequest struct {
    ClientTimestamp uint64
    SecurityCookie  uint32
    MTU             uint16
    ClientGUID      uint64
}

// ConnectionReply contains connection reply data
type ConnectionReply struct {
    ServerTimestamp uint64
    SecurityCookie  uint32
    MTU             uint16
    ServerGUID      uint64
    RequestAccepted bool
}

// PingData contains ping measurement data
type PingData struct {
    Timestamp     uint64
    ClientGUID    uint64
    SecurityCookie uint32
}

// PongData contains pong response data
type PongData struct {
    PingTimestamp  uint64
    ServerTimestamp uint64
    ServerGUID     uint64
    ServerName     string
}

// NetworkAddress represents a network address
type NetworkAddress struct {
    IP   [4]byte // IPv4 address
    Port uint16
    Type byte // 4 for IPv4, 6 for IPv6
}

// SystemAddress represents a system address in RakNet
type SystemAddress struct {
    Address NetworkAddress
    SystemIndex uint16
}

// ConnectionAttemptResult represents the result of a connection attempt
type ConnectionAttemptResult int

const (
    ConnectionAttemptStarted ConnectionAttemptResult = iota
    ConnectionAttemptInvalidParameter
    ConnectionAttemptAlreadyConnected
    ConnectionAttemptNoFreeSlots
    ConnectionAttemptDisconnected
    ConnectionAttemptConnectionInProgress
    ConnectionAttemptSecurityInitializationFailed
)

// String returns the string representation of ConnectionAttemptResult
func (r ConnectionAttemptResult) String() string {
    switch r {
    case ConnectionAttemptStarted:
        return "STARTED"
    case ConnectionAttemptInvalidParameter:
        return "INVALID_PARAMETER"
    case ConnectionAttemptAlreadyConnected:
        return "ALREADY_CONNECTED"
    case ConnectionAttemptNoFreeSlots:
        return "NO_FREE_SLOTS"
    case ConnectionAttemptDisconnected:
        return "DISCONNECTED"
    case ConnectionAttemptConnectionInProgress:
        return "CONNECTION_IN_PROGRESS"
    case ConnectionAttemptSecurityInitializationFailed:
        return "SECURITY_INITIALIZATION_FAILED"
    default:
        return "UNKNOWN"
    }
}

// PacketReliabilityMapping maps reliability types to their bit flags
var PacketReliabilityMapping = map[PacketReliability]byte{
    ReliabilityUnreliable:                      0,
    ReliabilityUnreliableSequenced:             1,
    ReliabilityReliable:                        2,
    ReliabilityReliableOrdered:                 3,
    ReliabilityReliableSequenced:               4,
    ReliabilityUnreliableWithAckReceipt:        5,
    ReliabilityReliableWithAckReceipt:          6,
    ReliabilityReliableOrderedWithAckReceipt:   7,
}

// ReversePacketReliabilityMapping maps bit flags back to reliability types
var ReversePacketReliabilityMapping = map[byte]PacketReliability{
    0: ReliabilityUnreliable,
    1: ReliabilityUnreliableSequenced,
    2: ReliabilityReliable,
    3: ReliabilityReliableOrdered,
    4: ReliabilityReliableSequenced,
    5: ReliabilityUnreliableWithAckReceipt,
    6: ReliabilityReliableWithAckReceipt,
    7: ReliabilityReliableOrderedWithAckReceipt,
}

// PacketFlags represents packet header flags
type PacketFlags byte

const (
    FlagValid             PacketFlags = 0x80
    FlagAck              PacketFlags = 0x40
    FlagNack             PacketFlags = 0x20
    FlagSplit            PacketFlags = 0x10
    FlagRejected         PacketFlags = 0x08
    FlagReserved1        PacketFlags = 0x04
    FlagReserved2        PacketFlags = 0x02
    FlagReserved3        PacketFlags = 0x01
)

// IsValid checks if the packet flag is valid
func (f PacketFlags) IsValid() bool {
    return f&FlagValid != 0
}

// IsAck checks if the packet is an ACK
func (f PacketFlags) IsAck() bool {
    return f&FlagAck != 0
}

// IsNack checks if the packet is a NACK
func (f PacketFlags) IsNack() bool {
    return f&FlagNack != 0
}

// HasSplit checks if the packet has split data
func (f PacketFlags) HasSplit() bool {
    return f&FlagSplit != 0
}

// ConnectionStatus represents detailed connection status
type ConnectionStatus struct {
    State               ConnectionState
    Metrics             ConnectionMetrics
    Statistics          NetworkStatistics
    RemoteAddress       *NetworkAddress
    LocalAddress        *NetworkAddress
    ConnectionTime      time.Time
    LastActivity        time.Time
    IsActive            bool
}

// NewConnectionStatus creates a new ConnectionStatus
func NewConnectionStatus() *ConnectionStatus {
    return &ConnectionStatus{
        State:          StateUnconnected,
        Metrics:        ConnectionMetrics{},
        Statistics:     NetworkStatistics{},
        RemoteAddress:  &NetworkAddress{},
        LocalAddress:   &NetworkAddress{},
        ConnectionTime: time.Now(),
        LastActivity:   time.Now(),
        IsActive:       false,
    }
}

// PacketBuffer represents a buffer for packet data
type PacketBuffer struct {
    Data      []byte
    Length    int
    Capacity  int
    ReadPos   int
    WritePos  int
}

// NewPacketBuffer creates a new packet buffer
func NewPacketBuffer(capacity int) *PacketBuffer {
    return &PacketBuffer{
        Data:     make([]byte, capacity),
        Capacity: capacity,
        ReadPos:  0,
        WritePos: 0,
    }
}

// Reset resets the packet buffer
func (b *PacketBuffer) Reset() {
    b.ReadPos = 0
    b.WritePos = 0
    b.Length = 0
}

// Write writes data to the buffer
func (b *PacketBuffer) Write(data []byte) error {
    if b.WritePos+len(data) > b.Capacity {
        return ErrBufferFull
    }
    copy(b.Data[b.WritePos:], data)
    b.WritePos += len(data)
    b.Length = b.WritePos - b.ReadPos
    return nil
}

// Read reads data from the buffer
func (b *PacketBuffer) Read(size int) ([]byte, error) {
    if b.ReadPos+size > b.WritePos {
        return nil, ErrBufferEmpty
    }
    data := b.Data[b.ReadPos : b.ReadPos+size]
    b.ReadPos += size
    b.Length = b.WritePos - b.ReadPos
    return data, nil
}

// Error types
var (
    ErrBufferFull  = &NetworkError{"buffer full"}
    ErrBufferEmpty = &NetworkError{"buffer empty"}
    ErrInvalidPacket = &NetworkError{"invalid packet"}
    ErrConnectionClosed = &NetworkError{"connection closed"}
    ErrTimeout     = &NetworkError{"timeout"}
)

// NetworkError represents a network-related error
type NetworkError struct {
    Message string
}

func (e *NetworkError) Error() string {
    return "network error: " + e.Message
}