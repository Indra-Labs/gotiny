package gotiny

import (
	"reflect"
	"unsafe"
)

type Encoder struct {
	buf     []byte // 编码目的数组 encoding destination array
	off     int
	boolPos int  // 下一次要设置的bool在buf中的下标,即buf[boolPos] The subscript of the bool to be set next time in buf, that is, buf[boolPos]
	boolBit byte // 下一次要设置的bool的buf[boolPos]中的bit位 The bit in buf[boolPos] of bool to be set next time

	engines []encEng
	length  int
}

func Marshal(is ...interface{}) []byte {
	return NewEncoderWithPtr(is...).Encode(is...)
}

// NewEncoderWithPtr 创建一个编码ps 指向类型的编码器 Create an encoder that encodes the type pointed to by ps
func NewEncoderWithPtr(ps ...interface{}) *Encoder {
	l := len(ps)
	engines := make([]encEng, l)
	for i := 0; i < l; i++ {
		rt := reflect.TypeOf(ps[i])
		if rt.Kind() != reflect.Ptr {
			panic("must a pointer type!")
		}
		engines[i] = getEncEngine(rt.Elem())
	}
	return &Encoder{
		length:  l,
		engines: engines,
	}
}

// NewEncoder 创建一个编码is 类型的编码器 Create an encoder that encodes the is type
func NewEncoder(is ...interface{}) *Encoder {
	l := len(is)
	engines := make([]encEng, l)
	for i := 0; i < l; i++ {
		engines[i] = getEncEngine(reflect.TypeOf(is[i]))
	}
	return &Encoder{
		length:  l,
		engines: engines,
	}
}

// NewEncoderWithType ...
func NewEncoderWithType(ts ...reflect.Type) *Encoder {
	l := len(ts)
	engines := make([]encEng, l)
	for i := 0; i < l; i++ {
		engines[i] = getEncEngine(ts[i])
	}
	return &Encoder{
		length:  l,
		engines: engines,
	}
}

// Encode 入参是要编码值的指针 The input parameter is a pointer to the value to be encoded
func (e *Encoder) Encode(is ...interface{}) []byte {
	engines := e.engines
	for i := 0; i < len(engines) && i < len(is); i++ {
		engines[i](e, (*[2]unsafe.Pointer)(unsafe.Pointer(&is[i]))[1])
	}
	return e.reset()
}

// EncodePtr 入参是要编码的值得unsafe.Pointer 指针 The input parameter is an unsafe.Pointer pointer worth encoding to
func (e *Encoder) EncodePtr(ps ...unsafe.Pointer) []byte {
	engines := e.engines
	for i := 0; i < len(engines) && i < len(ps); i++ {
		engines[i](e, ps[i])
	}
	return e.reset()
}

// EncodeValue vs 是持有要编码的值 is the value that holds the value to be encoded
func (e *Encoder) EncodeValue(vs ...reflect.Value) []byte {
	engines := e.engines
	for i := 0; i < len(engines) && i < len(vs); i++ {
		engines[i](e, getUnsafePointer(&vs[i]))
	}
	return e.reset()
}

// AppendTo 编码产生的数据将append到buf上 The data generated by encoding will be appended to buf
func (e *Encoder) AppendTo(buf []byte) {
	e.off = len(buf)
	e.buf = buf
}

func (e *Encoder) reset() []byte {
	buf := e.buf
	e.buf = buf[:e.off]
	e.boolBit = 0
	e.boolPos = 0
	return buf
}
