package student

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *GradeSet) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Grade":
			z.Grade, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Grade")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z GradeSet) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Grade"
	err = en.Append(0x81, 0xa5, 0x47, 0x72, 0x61, 0x64, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Grade)
	if err != nil {
		err = msgp.WrapError(err, "Grade")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z GradeSet) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "Grade"
	o = append(o, 0x81, 0xa5, 0x47, 0x72, 0x61, 0x64, 0x65)
	o = msgp.AppendString(o, z.Grade)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *GradeSet) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Grade":
			z.Grade, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Grade")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z GradeSet) Msgsize() (s int) {
	s = 1 + 6 + msgp.StringPrefixSize + len(z.Grade)
	return
}
