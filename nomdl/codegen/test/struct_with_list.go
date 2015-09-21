// This file was generated by nomdl/codegen.

package test

import (
	"github.com/attic-labs/noms/ref"
	"github.com/attic-labs/noms/types"
)

// StructWithList

type StructWithList struct {
	m types.Map
}

func NewStructWithList() StructWithList {
	return StructWithList{types.NewMap(
		types.NewString("$name"), types.NewString("StructWithList"),
		types.NewString("L"), types.NewList(),
		types.NewString("B"), types.Bool(false),
		types.NewString("S"), types.NewString(""),
		types.NewString("I"), types.Int64(0),
	)}
}

type StructWithListDef struct {
	L ListOfUInt8Def
	B bool
	S string
	I int64
}

func (def StructWithListDef) New() StructWithList {
	return StructWithList{
		types.NewMap(
			types.NewString("$name"), types.NewString("StructWithList"),
			types.NewString("L"), def.L.New().NomsValue(),
			types.NewString("B"), types.Bool(def.B),
			types.NewString("S"), types.NewString(def.S),
			types.NewString("I"), types.Int64(def.I),
		)}
}

func (self StructWithList) Def() StructWithListDef {
	return StructWithListDef{
		ListOfUInt8FromVal(self.m.Get(types.NewString("L"))).Def(),
		bool(self.m.Get(types.NewString("B")).(types.Bool)),
		self.m.Get(types.NewString("S")).(types.String).String(),
		int64(self.m.Get(types.NewString("I")).(types.Int64)),
	}
}

func StructWithListFromVal(val types.Value) StructWithList {
	// TODO: Validate here
	return StructWithList{val.(types.Map)}
}

func (self StructWithList) NomsValue() types.Value {
	return self.m
}

func (self StructWithList) Equals(other StructWithList) bool {
	return self.m.Equals(other.m)
}

func (self StructWithList) Ref() ref.Ref {
	return self.m.Ref()
}

func (self StructWithList) L() ListOfUInt8 {
	return ListOfUInt8FromVal(self.m.Get(types.NewString("L")))
}

func (self StructWithList) SetL(val ListOfUInt8) StructWithList {
	return StructWithList{self.m.Set(types.NewString("L"), val.NomsValue())}
}

func (self StructWithList) B() bool {
	return bool(self.m.Get(types.NewString("B")).(types.Bool))
}

func (self StructWithList) SetB(val bool) StructWithList {
	return StructWithList{self.m.Set(types.NewString("B"), types.Bool(val))}
}

func (self StructWithList) S() string {
	return self.m.Get(types.NewString("S")).(types.String).String()
}

func (self StructWithList) SetS(val string) StructWithList {
	return StructWithList{self.m.Set(types.NewString("S"), types.NewString(val))}
}

func (self StructWithList) I() int64 {
	return int64(self.m.Get(types.NewString("I")).(types.Int64))
}

func (self StructWithList) SetI(val int64) StructWithList {
	return StructWithList{self.m.Set(types.NewString("I"), types.Int64(val))}
}

// ListOfUInt8

type ListOfUInt8 struct {
	l types.List
}

func NewListOfUInt8() ListOfUInt8 {
	return ListOfUInt8{types.NewList()}
}

type ListOfUInt8Def []uint8

func (def ListOfUInt8Def) New() ListOfUInt8 {
	l := make([]types.Value, len(def))
	for i, d := range def {
		l[i] = types.UInt8(d)
	}
	return ListOfUInt8{types.NewList(l...)}
}

func (self ListOfUInt8) Def() ListOfUInt8Def {
	l := make([]uint8, self.Len())
	for i := uint64(0); i < self.Len(); i++ {
		l[i] = uint8(self.l.Get(i).(types.UInt8))
	}
	return l
}

func ListOfUInt8FromVal(val types.Value) ListOfUInt8 {
	// TODO: Validate here
	return ListOfUInt8{val.(types.List)}
}

func (self ListOfUInt8) NomsValue() types.Value {
	return self.l
}

func (l ListOfUInt8) Equals(p ListOfUInt8) bool {
	return l.l.Equals(p.l)
}

func (l ListOfUInt8) Ref() ref.Ref {
	return l.l.Ref()
}

func (l ListOfUInt8) Len() uint64 {
	return l.l.Len()
}

func (l ListOfUInt8) Empty() bool {
	return l.Len() == uint64(0)
}

func (self ListOfUInt8) Get(i uint64) uint8 {
	return uint8(self.l.Get(i).(types.UInt8))
}

func (l ListOfUInt8) Slice(idx uint64, end uint64) ListOfUInt8 {
	return ListOfUInt8{l.l.Slice(idx, end)}
}

func (self ListOfUInt8) Set(i uint64, val uint8) ListOfUInt8 {
	return ListOfUInt8{self.l.Set(i, types.UInt8(val))}
}

func (l ListOfUInt8) Append(v ...uint8) ListOfUInt8 {
	return ListOfUInt8{l.l.Append(l.fromElemSlice(v)...)}
}

func (l ListOfUInt8) Insert(idx uint64, v ...uint8) ListOfUInt8 {
	return ListOfUInt8{l.l.Insert(idx, l.fromElemSlice(v)...)}
}

func (l ListOfUInt8) Remove(idx uint64, end uint64) ListOfUInt8 {
	return ListOfUInt8{l.l.Remove(idx, end)}
}

func (l ListOfUInt8) RemoveAt(idx uint64) ListOfUInt8 {
	return ListOfUInt8{(l.l.RemoveAt(idx))}
}

func (l ListOfUInt8) fromElemSlice(p []uint8) []types.Value {
	r := make([]types.Value, len(p))
	for i, v := range p {
		r[i] = types.UInt8(v)
	}
	return r
}

type ListOfUInt8IterCallback func(v uint8) (stop bool)

func (l ListOfUInt8) Iter(cb ListOfUInt8IterCallback) {
	l.l.Iter(func(v types.Value) bool {
		return cb(uint8(v.(types.UInt8)))
	})
}

type ListOfUInt8IterAllCallback func(v uint8)

func (l ListOfUInt8) IterAll(cb ListOfUInt8IterAllCallback) {
	l.l.IterAll(func(v types.Value) {
		cb(uint8(v.(types.UInt8)))
	})
}

type ListOfUInt8FilterCallback func(v uint8) (keep bool)

func (l ListOfUInt8) Filter(cb ListOfUInt8FilterCallback) ListOfUInt8 {
	nl := NewListOfUInt8()
	l.IterAll(func(v uint8) {
		if cb(v) {
			nl = nl.Append(v)
		}
	})
	return nl
}
