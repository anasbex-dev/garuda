package raknet

const (
    ID_CONNECTED_PING          = 0x00
    ID_UNCONNECTED_PING        = 0x01
    ID_UNCONNECTED_PONG        = 0x1c
    ID_CONNECTED_PONG          = 0x03
    ID_OPEN_CONNECTION_REQUEST = 0x05
    ID_OPEN_CONNECTION_REPLY   = 0x06
    ID_CONNECTION_REQUEST      = 0x09
    ID_CONNECTION_REQUEST_ACCEPTED = 0x10
    ID_NEW_INCOMING_CONNECTION = 0x13
    ID_DISCONNECTION_NOTICE    = 0x15
    ID_INCOMPATIBLE_PROTOCOL   = 0x19
)

type Packet struct {
    ID   byte
    Data []byte
}

type UnconnectedPing struct {
    PingID    int64
    ClientGUID int64
    Magic     [16]byte
}

type UnconnectedPong struct {
    PingID    int64
    ServerGUID int64
    Magic     [16]byte
    MOTD      string
}

type OpenConnectionRequest1 struct {
    Magic       [16]byte
    Protocol    byte
    MTUSize     int
}

type OpenConnectionReply1 struct {
    Magic       [16]byte
    ServerGUID  int64
    UseSecurity bool
    MTUSize     int16
}

type OpenConnectionRequest2 struct {
    Magic          [16]byte
    ServerAddress  string
    MTUSize        int16
    ClientGUID     int64
}

type OpenConnectionReply2 struct {
    Magic          [16]byte
    ServerGUID     int64
    ClientAddress  string
    MTUSize        int16
    Encryption     bool
}