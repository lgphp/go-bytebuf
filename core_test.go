package bytebuf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCore1(t *testing.T) {
	body := []byte("i love you")
	buf, _ := NewByteBufWithCapacity(2, []byte{})
	assert.Equal(t, buf.capacity, 2)
	bodylen := len(body)
	buf.WriteInt32BE(int32(bodylen))
	buf.WriteBytes(body)
	println(buf.readerIndex, buf.writerIndex)

	packetlen, _ := buf.ReadInt32BE()

	println(packetlen)
	assert.Equal(t, packetlen, int32(10))
	println(buf.readerIndex, buf.writerIndex)
	println(buf.ReadableBytes())
	packetbytes := make([]byte, buf.ReadableBytes())
	//buf.ForceRelease()
	buf.ReadBytes(packetbytes)
	println(buf.readerIndex, buf.writerIndex)
	println(string(packetbytes))
	println(buf.refCount)

}

func TestCore2(t *testing.T) {

}
