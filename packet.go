package main

const (
	PROTOCOL_NAME    = "MQTT"
	PROTOCOL_VERSION = 4
	KEEP_ALIVE       = 120
	AUTH_COMMAND     = 0x10
	PUBLISH_COMMAND  = 0x30
	PUBACK_COMMAND   = 0x40

	SUBSCRIBE_COMMAND = 0x80

	PINGREQ     = 0xc0
	PINGRESP    = 0xd0
	RPC_PUBONL  = 0x20
	RPC_PUBOFFL = 0x30
	RPC_KICKOFF = 0x40
)

type Packet struct {
	fixHeader    uint8
	command      uint8
	remainLength uint32
	body         []byte
	bodyPos      uint32
}

func NewPacket(size uint32) *Packet {
	packet := &Packet{}
	packet.fixHeader = 0
	packet.command = 0
	packet.remainLength = 0
	packet.body = make([]byte, size)
	packet.bodyPos = 0
	return packet
}

func (self *Packet) readByte() byte {
	val := self.body[self.bodyPos]
	self.bodyPos += 1
	return val
}

func (self *Packet) writeByte(val byte) {
	self.body[self.bodyPos] = val
	self.bodyPos += 1
}

func (self *Packet) readInt16() uint16 {
	msb := self.body[self.bodyPos]
	self.bodyPos += 1
	lsb := self.body[self.bodyPos]
	self.bodyPos += 1

	return uint16(msb<<8 + lsb)
}

func (self *Packet) writeInt16(val uint16) {
	msb := byte((val & 0xFF00) >> 8)
	lsb := byte(val & 0x00FF)
	self.writeByte(msb)
	self.writeByte(lsb)
}

func (self *Packet) readBytes(length int) []byte {
	buf := make([]byte, length)
	copy(buf, self.body[self.bodyPos:])
	self.bodyPos += uint32(length)
	return buf
}

func (self *Packet) writeBytes(val []byte, length int) {
	copy(self.body[self.bodyPos:], val)
	self.bodyPos += uint32(length)
}

func (self *Packet) readString() string {
	length := self.readInt16()
	buf := self.readBytes(int(length))
	return string(buf)
}

func (self *Packet) writeString(val string, length int) {
	self.writeInt16(uint16(length))
	self.writeBytes([]byte(val), length)
}

func ReceivePacket(conn *Conn) (*Packet, error) {

	packet := &Packet{}

	buf1 := make([]byte, 1)

	//read head
	_, err := conn.Read(buf1)
	if err != nil {
		return nil, err
	}
	packet.fixHeader = buf1[0]
	packet.command = buf1[0] & 0xF0

	//read remainLength
	remainLength := 0
	multiplier := 1
	for {
		_, err := conn.Read(buf1)
		if err != nil {
			return nil, err
		}
		remainLength += int(buf1[0]&127) * multiplier
		multiplier *= 128
		if buf1[0]&128 == 0 {
			break
		}
	}
	packet.remainLength = uint32(remainLength)
	if remainLength > 0 {
		packet.body = make([]byte, remainLength)
		haveRead := 0
		for haveRead < remainLength {
			curLen, err := conn.Read(packet.body[haveRead:])
			if nil != err {
				return nil, err
			}
			haveRead += curLen
		}
	}
	return packet, nil

}

func SendPacket(conn *Conn, packet *Packet) error {
	buf := make([]byte, 1)
	buf[0] = packet.fixHeader

	length := packet.remainLength

	if length == 0 {
		buf = append(buf, 0)
	} else {

		for {
			digit := length % 128
			length = length / 128
			if length > 0 {
				digit = digit | 0x80
			}

			buf = append(buf, byte(digit))

			if length <= 0 {
				break
			}
		}

		buf = append(buf, packet.body...)
	}

	haveWrite := 0
	total := len(buf)
	for haveWrite < total {
		curLen, err := conn.Write(buf[haveWrite:])
		if nil != err {
			return err
		}
		haveWrite += curLen
	}

	return nil

}

func GainAuthPacket(clientId, name, passwd string) *Packet {
	remainLength := 2 + len(PROTOCOL_NAME) +
		1 + //proto version
		1 + //proto flag
		2 + //keep alive time
		2 + len(clientId) +
		2 + len(name) +
		2 + len(passwd)

	flag := 1<<7 | 1<<6

	packet := NewPacket(uint32(remainLength))
	packet.remainLength = uint32(remainLength)
	packet.command = AUTH_COMMAND
	packet.fixHeader = AUTH_COMMAND

	packet.writeString(PROTOCOL_NAME, len(PROTOCOL_NAME))
	packet.writeByte(PROTOCOL_VERSION)
	packet.writeByte(byte(flag))
	packet.writeInt16(KEEP_ALIVE)
	packet.writeString(clientId, len(clientId))
	packet.writeString(name, len(name))
	packet.writeString(passwd, len(passwd))

	return packet

}

func GainInstantMsgPacket(publishId string, topic string, message string, clientIds []string) *Packet {

	clientCount := len(clientIds)

	remainLength := 2 + len(publishId) +
		2 + len(topic) +
		2 + len(message) +
		2 //client id count (uint16)

	for _, clientId := range clientIds {
		remainLength += 2 + len(clientId)
	}

	packet := NewPacket(uint32(remainLength))
	packet.remainLength = uint32(remainLength)
	packet.command = RPC_PUBONL
	packet.fixHeader = packet.command

	packet.writeString(publishId, len(publishId))
	packet.writeString(topic, len(topic))
	packet.writeString(message, len(message))
	packet.writeInt16(uint16(clientCount))
	for _, clientId := range clientIds {
		packet.writeString(clientId, len(clientId))
	}
	return packet
}

func GainOfflineMsgPacket(clientId string, topic string, msgList []map[string]string) *Packet {

	msgCount := len(msgList)

	remainLength := 2 + len(clientId) +
		2 + len(topic) +
		2 //msg count (uint16)

	for _, msg := range msgList {
		remainLength += 2 + len(msg["publishId"]) +
			2 + len(msg["message"])
	}

	packet := NewPacket(uint32(remainLength))
	packet.remainLength = uint32(remainLength)
	packet.command = RPC_PUBOFFL
	packet.fixHeader = packet.command

	packet.writeString(clientId, len(clientId))
	packet.writeString(topic, len(topic))
	packet.writeInt16(uint16(msgCount))
	for _, msg := range msgList {
		packet.writeString(msg["publishId"], len(msg["publishId"]))
		packet.writeString(msg["message"], len(msg["message"]))
	}
	return packet
}

func GainKickOffPacket(clientId, timeSerie string) *Packet {
	remainLength := 2 + len(clientId) +
		2 + len(timeSerie)

	packet := NewPacket(uint32(remainLength))
	packet.remainLength = uint32(remainLength)
	packet.command = RPC_KICKOFF
	packet.fixHeader = packet.command

	packet.writeString(clientId, len(clientId))
	packet.writeString(timeSerie, len(timeSerie))

	return packet
}

func GainSubscribePacket(topic string) *Packet {
	msgid := 100
	qos := 1
	remainLength := 2 + //msgid
		2 + len(topic) + //topic
		1 //qos

	packet := NewPacket(uint32(remainLength))
	packet.remainLength = uint32(remainLength)
	packet.command = SUBSCRIBE_COMMAND
	packet.fixHeader = packet.command | 3

	packet.writeInt16(uint16(msgid))
	packet.writeString(topic, len(topic))
	packet.writeByte(byte(qos))

	return packet
}
func GainPublishAckPacket(msgId int) *Packet {
	remainLength := 2 //msgid

	packet := NewPacket(uint32(remainLength))
	packet.remainLength = uint32(remainLength)
	packet.command = PUBACK_COMMAND
	packet.fixHeader = packet.command

	packet.writeInt16(uint16(msgId))

	return packet
}

func GainPingPacket() *Packet {
	packet := NewPacket(0)
	packet.remainLength = 0
	packet.command = PINGREQ
	packet.fixHeader = packet.command

	return packet
}
