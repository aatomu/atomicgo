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
	BadAuth        int32    = -1
	Response       SendType = 0
	ExecuteCommand SendType = 2
	AuthSuccess    SendType = 2
	ServerAuth     SendType = 3
)

var uniqueID = 1

// Rcon処理用のStruct
type Rcon struct {
	conn net.Conn
	pass string
}

// Rcon起動
func Login(address, password string) (c *Rcon, err error) {
	// tcp送信
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return nil, err
	}
	// 保存
	c = &Rcon{conn: conn, pass: password}
	// 認証
	conn.Write(format(password, ServerAuth))
	// 読み込み
	r := c.result()
	if r.ID < 1 {
		return nil, fmt.Errorf("failed rcon server auth")
	}
	return
}

// コマンドの送信
func (c *Rcon) SendCommand(cmd string) (result string) {
	// 送信
	c.conn.Write(format(cmd, ExecuteCommand))
	// 読み込み
	r := c.result()
	return r.Body
}

func (c *Rcon) Close() error {
	return c.conn.Close()
}

// 整形
func format(cmd string, sendType SendType) []byte {
	body := []byte(cmd)
	size := int32(4 + 4 + len([]byte(body)) + 2)
	uniqueID++
	var id int32 = int32(uniqueID)

	p := packet{}
	p.Write(size)
	p.Write(id)
	p.Write(sendType)
	p.Write(body)
	p.Write([]byte{0x0, 0x0})
	return p.buffer.Bytes()
}

// 実行結果を入手
func (c *Rcon) result() (result packetBody) {
	b := make([]byte, 4096)
	c.conn.Read(b)
	buf := bytes.NewBuffer(b)
	p := packet{buffer: buf}

	p.Read(&result.Size)
	p.Read(&result.ID)
	p.Read(&result.Type)
	body := buf.Bytes()
	result.Body = string(body[:result.Size-4-4-2]) // 全体のサイズ-ID-Type-Null文字
	return
}

type packet struct {
	buffer *bytes.Buffer
}

type packetBody struct {
	Size int32
	ID   int32
	Type int32
	Body string
}

func (p *packet) Write(v interface{}) {
	if p.buffer == nil {
		p.buffer = new(bytes.Buffer)
	}
	binary.Write(p.buffer, binary.LittleEndian, v)
}

func (p *packet) Read(v interface{}) {
	binary.Read(p.buffer, binary.LittleEndian, v)
}
