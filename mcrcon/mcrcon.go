package mcrcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// Rcon送信タイプ
type SendType int32

const (
	BadAuth        SendType = -1
	Response       SendType = 0
	ExecuteCommand SendType = 2
	AuthSuccess    SendType = 2
	ServerAuth     SendType = 3
)

// Rcon処理用のStruct
type Rcon struct {
	conn     net.Conn
	pass     string
	uniqueID int32
}

// Rcon起動
func Login(address, password string) (rcon *Rcon, err error) {
	// tcp送信
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return nil, err
	}
	// 保存
	rcon = &Rcon{
		conn:     conn,
		pass:     password,
		uniqueID: 0,
	}

	// 送信
	res, err := rcon.Send(ServerAuth, password)
	if res.Type != AuthSuccess {
		return nil, fmt.Errorf("failed rcon server auth")
	}
	return
}

// コマンドの送信
func (c *Rcon) SendCommand(cmd string) (result *RconRes, err error) {
	return c.Send(ExecuteCommand, cmd)
}

func (c *Rcon) Close() error {
	return c.conn.Close()
}

type RconRes struct {
	Size int32
	ID   int32
	Type SendType
	Body []byte
}

func (rcon *Rcon) Send(sendType SendType, body string) (res *RconRes, err error) {
	// Fromat
	bodyBytes := []byte(body)
	size := int32(4 + 4 + len(bodyBytes) + 2)
	packetID := rcon.uniqueID
	rcon.uniqueID++

	sendPacket := packet{}
	sendPacket.write(size)
	sendPacket.write(packetID)
	sendPacket.write(sendType)
	sendPacket.write(bodyBytes)
	sendPacket.write([]byte{0x0, 0x0})
	// Send
	_, err = rcon.conn.Write(sendPacket.buffer.Bytes())
	if err != nil {
		return nil, err
	}

	// Read
	rcon.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	b := make([]byte, 4096)
	n, err := rcon.conn.Read(b)
	if err != nil {
		return nil, err
	}

	readBuf := bytes.NewBuffer(b[:n])
	readPacket := packet{buffer: readBuf}
	res = &RconRes{}
	readPacket.read(&res.Size)
	readPacket.read(&res.ID)
	readPacket.read(&res.Type)
	res.Body = readBuf.Bytes()[:res.Size-4-4-2] // 全体のサイズ-ID-Type-Null文字
	return
}

type packet struct {
	buffer *bytes.Buffer
}

func (p *packet) write(v interface{}) {
	if p.buffer == nil {
		p.buffer = new(bytes.Buffer)
	}
	binary.Write(p.buffer, binary.LittleEndian, v)
}

func (p *packet) read(v interface{}) {
	binary.Read(p.buffer, binary.LittleEndian, v)
}
