// This file was generated by nomdl/codegen.

package test

import (
	"github.com/attic-labs/noms/ref"
	"github.com/attic-labs/noms/types"
)

// SetOfBool

type SetOfBool struct {
	s types.Set
}

func NewSetOfBool() SetOfBool {
	return SetOfBool{types.NewSet()}
}

type SetOfBoolDef map[bool]bool

func (def SetOfBoolDef) New() SetOfBool {
	l := make([]types.Value, len(def))
	i := 0
	for d, _ := range def {
		l[i] = types.Bool(d)
		i++
	}
	return SetOfBool{types.NewSet(l...)}
}

func (s SetOfBool) Def() SetOfBoolDef {
	def := make(map[bool]bool, s.Len())
	s.s.Iter(func(v types.Value) bool {
		def[bool(v.(types.Bool))] = true
		return false
	})
	return def
}

func SetOfBoolFromVal(p types.Value) SetOfBool {
	return SetOfBool{p.(types.Set)}
}

func (s SetOfBool) NomsValue() types.Value {
	return s.s
}

func (s SetOfBool) Equals(p SetOfBool) bool {
	return s.s.Equals(p.s)
}

func (s SetOfBool) Ref() ref.Ref {
	return s.s.Ref()
}

func (s SetOfBool) Empty() bool {
	return s.s.Empty()
}

func (s SetOfBool) Len() uint64 {
	return s.s.Len()
}

func (s SetOfBool) Has(p bool) bool {
	return s.s.Has(types.Bool(p))
}

type SetOfBoolIterCallback func(p bool) (stop bool)

func (s SetOfBool) Iter(cb SetOfBoolIterCallback) {
	s.s.Iter(func(v types.Value) bool {
		return cb(bool(v.(types.Bool)))
	})
}

type SetOfBoolIterAllCallback func(p bool)

func (s SetOfBool) IterAll(cb SetOfBoolIterAllCallback) {
	s.s.IterAll(func(v types.Value) {
		cb(bool(v.(types.Bool)))
	})
}

type SetOfBoolFilterCallback func(p bool) (keep bool)

func (s SetOfBool) Filter(cb SetOfBoolFilterCallback) SetOfBool {
	ns := NewSetOfBool()
	s.IterAll(func(v bool) {
		if cb(v) {
			ns = ns.Insert(v)
		}
	})
	return ns
}

func (s SetOfBool) Insert(p ...bool) SetOfBool {
	return SetOfBool{s.s.Insert(s.fromElemSlice(p)...)}
}

func (s SetOfBool) Remove(p ...bool) SetOfBool {
	return SetOfBool{s.s.Remove(s.fromElemSlice(p)...)}
}

func (s SetOfBool) Union(others ...SetOfBool) SetOfBool {
	return SetOfBool{s.s.Union(s.fromStructSlice(others)...)}
}

func (s SetOfBool) Subtract(others ...SetOfBool) SetOfBool {
	return SetOfBool{s.s.Subtract(s.fromStructSlice(others)...)}
}

func (s SetOfBool) Any() bool {
	return bool(s.s.Any().(types.Bool))
}

func (s SetOfBool) fromStructSlice(p []SetOfBool) []types.Set {
	r := make([]types.Set, len(p))
	for i, v := range p {
		r[i] = v.s
	}
	return r
}

func (s SetOfBool) fromElemSlice(p []bool) []types.Value {
	r := make([]types.Value, len(p))
	for i, v := range p {
		r[i] = types.Bool(v)
	}
	return r
}
