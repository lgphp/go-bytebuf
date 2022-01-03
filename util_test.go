package bytebuf

import (
	"testing"
)

func Test_Read(t *testing.T) {
	str := "中华人民共和国"
	buf, _ := NewByteBufWithCapacity(0)
	_ = buf.WriteStringWithByteLength(str)
	re := buf.ReadStringWithByteLength()
	println(re)

}
