package main

import (
	"fmt"
	"net"
	"time"
)

type Conn struct {
	sock net.Conn
}

func (p *Conn) SetTimeout(timeout int) {
	p.sock.SetReadDeadline(time.Now().Add(time.Second *
		time.Duration(timeout)))
}

func (p *Conn) Read(buf []byte) (n int, err error) {
	return p.sock.Read(buf)
}
func (p *Conn) Write(buf []byte) (n int, err error) {
	return p.sock.Write(buf)

}

type BrokerConn struct {
	addr     string
	port     int
	Using    bool
	conn     *Conn
	timeout  int
	writeErr bool
}

func NewConn(addr string, port int) *Conn {
	conn := &Conn{}

	address := fmt.Sprintf("%s:%d", addr, port)
	duration := time.Second * time.Duration(3)
	sock, err := net.DialTimeout("tcp", address, duration)
	if err != nil {
		fmt.Printf("conncet to <%s> failed, %v\n", address, err)
		return nil
	}
	conn.sock = sock
	return conn
}

func NewBrokerConn(addr string, port int, timeout int) *BrokerConn {
	client := &BrokerConn{}
	client.addr = addr
	client.port = port
	client.timeout = timeout
	if client.conn = NewConn(addr, port); client.conn == nil {
		return nil
	}

	client.writeErr = false
	return client
}

func (self *BrokerConn) IsActive() bool {
	if self.writeErr {
		return false
	}
	return self.SendPing()
}

func (self *BrokerConn) SendPing() bool {
	packet := GainPingPacket()
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
		return false
	}
	if self.timeout > 0 {
		self.conn.SetTimeout(self.timeout)
	}
	buf := make([]byte, 2)
	_, err = self.conn.Read(buf)
	if nil == err {
		return buf[0] == PINGRESP
	}

	return false
}

func (self *BrokerConn) PublishInstantMsg(publishId string, topic string, message string, clientIds []string) error {
	packet := GainInstantMsgPacket(publishId, topic, message, clientIds)
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
	}
	return err
}

func (self *BrokerConn) PublishOfflineMsg(clientId string, topic string, msgList []map[string]string) error {
	packet := GainOfflineMsgPacket(clientId, topic, msgList)
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
	}
	return err
}

func (self *BrokerConn) KickClientOff(clientId, timeSerie string) error {
	packet := GainKickOffPacket(clientId, timeSerie)
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
	}
	return err
}

func (self *BrokerConn) Auth(clientId, name, passwd string) error {
	packet := GainAuthPacket(clientId, name, passwd)
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
	}
	if self.timeout > 0 {
		self.conn.SetTimeout(self.timeout)
	}
	buf := make([]byte, 4)
	_, err = self.conn.Read(buf)
	return err
}

func (self *BrokerConn) Subscribe(topic string) error {
	packet := GainSubscribePacket(topic)
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
	}
	return err
}

func (self *BrokerConn) PublishAck(msgId int) error {
	packet := GainPublishAckPacket(msgId)
	err := SendPacket(self.conn, packet)
	if nil != err {
		self.writeErr = true
	}
	return err
}

func (self *BrokerConn) Release() {
	self.Using = false
}
