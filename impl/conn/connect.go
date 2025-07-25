package conn

import (
	"bytes"
	"compress/zlib"
	"crypto/cipher"
	"io"
	"log"
	"net"

	"github.com/Relixik/minecraft-server/apis/rand"
	"github.com/Relixik/minecraft-server/impl/base"
	"github.com/Relixik/minecraft-server/impl/conn/crypto"
)

type connection struct {
	new bool
	tcp *net.TCPConn

	state base.PacketState

	certify Certify
	compact Compact
}

func NewConnection(conn *net.TCPConn) base.Connection {
	return &connection{
		new: true,
		tcp: conn,

		certify: Certify{},
		compact: Compact{},
	}
}

func (c *connection) Address() net.Addr {
	return c.tcp.RemoteAddr()
}

func (c *connection) GetState() base.PacketState {
	return c.state
}

func (c *connection) SetState(state base.PacketState) {
	c.state = state
}

type Certify struct {
	name string

	used bool
	data []byte

	encrypt cipher.Stream
	decrypt cipher.Stream
}

func (c *connection) Encrypt(data []byte) (output []byte) {
	if !c.certify.used {
		return data
	}

	output = make([]byte, len(data))
	c.certify.encrypt.XORKeyStream(output, data)

	return
}

func (c *connection) Decrypt(data []byte) (output []byte) {
	if !c.certify.used {
		return data
	}

	output = make([]byte, len(data))
	c.certify.decrypt.XORKeyStream(output, data)

	return
}

func (c *connection) CertifyName() string {
	return c.certify.name
}

func (c *connection) CertifyData() []byte {
	return c.certify.data
}

func (c *connection) CertifyUpdate(secret []byte) {
	encrypt, decrypt, err := crypto.NewEncryptAndDecrypt(secret)

	if err != nil {
		log.Printf("CRITICAL: Failed to enable encryption for user %s: %v", c.CertifyName(), err)
		// Close the connection since encryption is mandatory
		c.tcp.Close()
		return
	}

	c.certify.encrypt = encrypt
	c.certify.decrypt = decrypt
	c.certify.used = true
	c.certify.data = secret
}

func (c *connection) CertifyValues(name string) {
	c.certify.name = name
	c.certify.data = rand.RandomByteArray(4)
}

type Compact struct {
	used bool
	size int32
}

func (c *connection) Deflate(data []byte) (output []byte) {
	if !c.compact.used {
		return data
	}

	var out bytes.Buffer

	writer, _ := zlib.NewWriterLevel(&out, zlib.BestCompression)
	_, _ = writer.Write(data)
	_ = writer.Close()

	output = out.Bytes()

	return
}

func (c *connection) Inflate(data []byte) (output []byte) {
	if !c.compact.used {
		return data
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		log.Printf("WARNING: Failed to decompress data for user %s: %v - returning original data", c.CertifyName(), err)
		return data // Return original data as fallback
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, reader)
	if err != nil {
		log.Printf("WARNING: Failed to read decompressed data for user %s: %v - returning original data", c.CertifyName(), err)
		return data // Return original data as fallback
	}

	output = out.Bytes()

	return
}

func (c *connection) Pull(data []byte) (len int, err error) {
	len, err = c.tcp.Read(data)
	return
}

func (c *connection) Push(data []byte) (len int, err error) {
	len, err = c.tcp.Write(data)
	return
}

func (c *connection) Stop() (err error) {
	err = c.tcp.Close()
	return
}

func (c *connection) SendPacket(packet base.PacketO) {
	bufO := NewBuffer()
	temp := NewBuffer()

	// write buffer
	bufO.PushVrI(packet.UUID())
	packet.Push(bufO, c)

	temp.PushVrI(bufO.Len())
	temp.PushUAS(bufO.UAS(), false)

	_, _ = c.tcp.Write(c.Encrypt(temp.UAS()))
}
