package bytebuf

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
)

// Global parameters, users can modify accordingly
var (
	// ChunkSize how much byte to allocate if there were not enough room
	ChunkSize = 1024
)

// Errors
var (
	ErrOutOfBound = errors.New("index is out of bound")
)

// ByteIterator type used for byte iteration
type ByteIterator func(byte) bool

func (b *ByteBuf) SkipBytes(length int) {
	b.readerIndex = b.readerIndex + length
}

func (b *ByteBuf) EnsureWritable(size int) {
	b.EnsureCapacity(b.writerIndex + size)
	//if ((b.capacity - b.writerIndex )> size){
	//	return
	//}
	//oldWritable := b.capacity - b.writerIndex
	//offset := size  - oldWritable
	//capacity := b.capacity + offset
	//newBuf := make([]byte, capacity)
	//copy(newBuf, b.buf)
	//b.buf = nil
	//b.buf = newBuf
	//b.capacity = capacity
}

// EnsureCapacity allocate more bytes to ensure the capacity
func (b *ByteBuf) EnsureCapacity(size int) {
	if size <= b.capacity {
		return
	}
	capacity := ((size-1)/ChunkSize + 1) * ChunkSize
	newBuf := make([]byte, capacity)
	copy(newBuf, b.buf)
	b.buf = nil
	b.buf = newBuf
	b.capacity = capacity
}

// Index index a byte slice inside buffer, Index(p), Index(p, start), Index(p, start, end)
func (b *ByteBuf) Index(p []byte, indexes ...int) int {
	if len(indexes) >= 2 {
		if indexes[1] > indexes[0] && indexes[1] < b.capacity && indexes[0] > 0 {
			return bytes.Index(b.buf[indexes[0]:indexes[1]], p)
		}
		return -1
	} else if len(indexes) >= 1 {
		if indexes[0] > 0 && indexes[0] < b.capacity {
			return bytes.Index(b.buf[indexes[0]:], p)
		}
		return -1
	}
	return bytes.Index(b.buf, p)
}

// Equal if current ByteBuf is equal to another ByteBuf, compared by underlying byte slice
func (b *ByteBuf) Equal(ob *ByteBuf) bool {
	return bytes.Equal(ob.buf, b.buf)
}

// DiscardReadBytes discard bytes that are read, adjust readerIndex/writerIndex accordingly
func (b *ByteBuf) DiscardReadBytes() {
	b.buf = b.buf[b.readerIndex:]
	b.writerIndex -= b.readerIndex
	b.readerIndex = 0
}

// Copy deep copy to create an brand new ByteBuf
func (b *ByteBuf) Copy() *ByteBuf {
	p := make([]byte, len(b.buf))
	copy(p, b.buf)
	return &ByteBuf{
		buf: p,

		capacity: b.capacity,

		readerIndex:  b.readerIndex,
		readerMarker: b.readerMarker,
		writerIndex:  b.writerIndex,
		writerMarker: b.writerMarker,
	}
}

// ForEachByte iterate through readable bytes, ForEachByte(iterator, start), ForEachByte(iterator, start, end)
func (b *ByteBuf) ForEachByte(iterator ByteIterator, indexes ...int) int {
	start := b.readerIndex
	end := b.writerIndex
	if len(indexes) >= 1 {
		start = indexes[0]
	}
	if len(indexes) >= 2 && indexes[1] < end {
		end = indexes[1]
	}

	if start > end {
		return 0
	}

	if end > b.capacity {
		end = b.capacity
	}

	count := 0
	for ; start < end; start++ {
		if !iterator(b.buf[start]) {
			break
		}
		count++
	}
	return count
}

// NewReader create ByteBuf from io.Reader
func NewReader(reader io.Reader) (*ByteBuf, error) {
	b, err := NewByteBuf()
	if err != nil {
		return nil, err
	}
	return b, b.ReadFrom(reader)
}

// String buf to string
func (b *ByteBuf) String() string {
	return string(b.AvailableBytes())
}

// 转换成base64
func (b *ByteBuf) Base64String() string {
	return base64.StdEncoding.EncodeToString(b.AvailableBytes())
}

// 转换成16进制
func (b *ByteBuf) HexString() string {
	return hex.EncodeToString(b.AvailableBytes())
}

func (b *ByteBuf) readString(len int) string {
	buffer := make([]byte, len)
	_, _ = b.ReadBytes(buffer)
	return string(buffer)
}

func (b *ByteBuf) ReadStringWithByteLength() string {
	l, _ := b.ReadByte()
	return b.readString(int(l))
}

func (b *ByteBuf) ReadStringWithU16Length() string {
	l, _ := b.ReadUInt16BE()
	return b.readString(int(l))
}

func (b *ByteBuf) ReadStringWithU32Length() string {
	l, _ := b.ReadUInt32BE()
	return b.readString(int(l))
}

// 传指针进来
func (b *ByteBuf) ReadObjectWithU32Length(v interface{}) {
	l, _ := b.ReadUInt32BE()
	buffer := make([]byte, l)
	_, _ = b.ReadBytes(buffer)
	_ = json.Unmarshal(buffer, v)
}

func (b *ByteBuf) writeString(bytes []byte) {
	if len(bytes) > 0 {
		_ = b.WriteBytes(bytes)
	}
}

func (b *ByteBuf) WriteStringWithByteLength(str string) error {
	bs := []byte(str)
	if len(bs) > math.MaxUint8 {
		return errors.New("str length exceed MaxUint8")
	}
	_ = b.WriteByte(byte(len(bs)))
	b.writeString(bs)
	return nil
}

func (b *ByteBuf) WriteStringWithU16Length(str string) error {
	bs := []byte(str)
	if len(bs) > math.MaxUint16 {
		return errors.New("str length exceed MaxUint16")
	}
	_ = b.WriteUInt16BE(uint16(len(bs)))
	b.writeString(bs)
	return nil
}

func (b *ByteBuf) WriteStringWithU32Length(str string) error {
	bs := []byte(str)
	if len(bs) > math.MaxUint32 {
		return errors.New("str length exceed MaxUint32")
	}
	_ = b.WriteUInt32BE(uint32(len(bs)))
	b.writeString(bs)
	return nil
}

func (b *ByteBuf) WriteObjectWithU32Length(v interface{}) error {
	mbs, e := json.Marshal(v)
	if e != nil {
		return e
	}
	if len(mbs) > math.MaxUint32 {
		return errors.New("v length exceed MaxUint32")
	}
	e = b.WriteUInt32BE(uint32(len(mbs)))
	if e != nil {
		return e
	}
	e = b.WriteBytes(mbs)
	return e
}
