// Copyright 2022 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tree

import (
	"bytes"
	"context"
	"errors"
	"io"
)

type DiffType byte

const (
	AddedDiff    DiffType = 0
	ModifiedDiff DiffType = 1
	RemovedDiff  DiffType = 2
)

type Diff struct {
	Key      Item
	From, To Item
	Type     DiffType
}

type DiffFn func(context.Context, Diff) error

// Differ computes the diff between two prolly trees.
// If `considerAllRowsModified` is true, it will consider every leaf to be modified and generate a diff for every leaf. (This
// is useful in cases where the schema has changed and we want to consider a leaf changed even if the byte representation
// of the leaf is the same.
type Differ[K ~[]byte, O Ordering[K]] struct {
	from, to                *cursor
	fromStop, toStop        *cursor
	order                   O
	considerAllRowsModified bool
}

func DifferFromRoots[K ~[]byte, O Ordering[K]](
	ctx context.Context,
	fromNs NodeStore, toNs NodeStore,
	from, to Node,
	order O,
	considerAllRowsModified bool,
) (Differ[K, O], error) {
	var fc, tc *cursor
	var err error

	if !from.empty() {
		fc, err = newCursorAtStart(ctx, fromNs, from)
		if err != nil {
			return Differ[K, O]{}, err
		}
	} else {
		fc = &cursor{}
	}

	if !to.empty() {
		tc, err = newCursorAtStart(ctx, toNs, to)
		if err != nil {
			return Differ[K, O]{}, err
		}
	} else {
		tc = &cursor{}
	}

	fs, err := newCursorPastEnd(ctx, fromNs, from)
	if err != nil {
		return Differ[K, O]{}, err
	}

	ts, err := newCursorPastEnd(ctx, toNs, to)
	if err != nil {
		return Differ[K, O]{}, err
	}

	return Differ[K, O]{
		from:                    fc,
		to:                      tc,
		fromStop:                fs,
		toStop:                  ts,
		order:                   order,
		considerAllRowsModified: considerAllRowsModified,
	}, nil
}

func DifferFromCursors[K ~[]byte, O Ordering[K]](
	ctx context.Context,
	fromRoot, toRoot Node,
	findStart, findStop SearchFn,
	fromStore, toStore NodeStore,
	order O,
) (Differ[K, O], error) {
	fromStart, err := newCursorFromSearchFn(ctx, fromStore, fromRoot, findStart)
	if err != nil {
		return Differ[K, O]{}, err
	}
	toStart, err := newCursorFromSearchFn(ctx, toStore, toRoot, findStart)
	if err != nil {
		return Differ[K, O]{}, err
	}
	fromStop, err := newCursorFromSearchFn(ctx, fromStore, fromRoot, findStop)
	if err != nil {
		return Differ[K, O]{}, err
	}
	toStop, err := newCursorFromSearchFn(ctx, toStore, toRoot, findStop)
	if err != nil {
		return Differ[K, O]{}, err
	}
	return Differ[K, O]{
		from:     fromStart,
		to:       toStart,
		fromStop: fromStop,
		toStop:   toStop,
		order:    order,
	}, nil
}

func (td Differ[K, O]) Next(ctx context.Context) (diff Diff, err error) {
	for td.from.Valid() && td.from.compare(td.fromStop) < 0 && td.to.Valid() && td.to.compare(td.toStop) < 0 {

		f := td.from.CurrentKey()
		t := td.to.CurrentKey()
		cmp := td.order.Compare(K(f), K(t))

		switch {
		case cmp < 0:
			return sendRemoved(ctx, td.from)

		case cmp > 0:
			return sendAdded(ctx, td.to)

		case cmp == 0:
			// If the cursor schema has changed, then all rows should be considered modified.
			// If the cursor schema hasn't changed, rows are modified iff their bytes have changed.
			if td.considerAllRowsModified || !equalcursorValues(td.from, td.to) {
				return sendModified(ctx, td.from, td.to)
			}

			// advance both cursors since we have already determined that they are equal. This needs to be done because
			// skipCommon will not advance the cursors if they are equal in a collation sensitive comparison but differ
			// in a byte comparison.
			if err = td.from.advance(ctx); err != nil {
				return Diff{}, err
			}
			if err = td.to.advance(ctx); err != nil {
				return Diff{}, err
			}

			// seek ahead to the next diff and loop again
			if err = skipCommon(ctx, td.from, td.to); err != nil {
				return Diff{}, err
			}
		}
	}

	if td.from.Valid() && td.from.compare(td.fromStop) < 0 {
		return sendRemoved(ctx, td.from)
	}
	if td.to.Valid() && td.to.compare(td.toStop) < 0 {
		return sendAdded(ctx, td.to)
	}

	return Diff{}, io.EOF
}

// NM4. gdoc.
func ChunkAddressDiffOrderedTrees[K, V ~[]byte, O Ordering[K]](
	ctx context.Context,
	from, to StaticMap[K, V, O],
	cb DiffFn,
) error {
	differ, layerDepth, err := AddressDifferFromRoots[K](ctx, from.NodeStore, to.NodeStore, from.Root, to.Root, from.Order, 1000)
	if err != nil {
		return err
	}
	for {
		for {
			var diff Diff
			if diff, err = differ.Next(ctx); err != nil {
				break
			}

			if err = cb(ctx, diff); err != nil {
				break
			}
		}
		if layerDepth > 0 {
			differ, layerDepth, err = AddressDifferFromRoots[K](ctx, from.NodeStore, to.NodeStore, from.Root, to.Root, from.Order, layerDepth)
			if err != nil {
				return err
			}
		} else {
			break
		}
	}
	return err
}

var ErrShallowTree = errors.New("tree too shallow") // NM4 We can do better than this. TBD.
func AddressDifferFromRoots[K ~[]byte, O Ordering[K]](
	ctx context.Context,
	fromNs NodeStore, toNs NodeStore,
	from, to Node,
	order O,
	desiredLayer int,
) (Differ[K, O], int, error) {
	orig, err := DifferFromRoots[K, O](ctx, fromNs, toNs, from, to, order, false)
	if err != nil {
		return orig, -1, err
	}

	depthActual := depth(orig.from)
	if depthActual < desiredLayer {
		desiredLayer = depthActual
	}

	for desiredLayer <= depthActual {
		orig.from = orig.from.parent
		orig.to = orig.to.parent
		orig.fromStop = orig.fromStop.parent
		orig.toStop = orig.toStop.parent

		if orig.from == nil || orig.to == nil || orig.fromStop == nil || orig.toStop == nil {
			return orig, -1, ErrShallowTree
		}
		depthActual--
	}

	return orig, depthActual, err
}

func depth(n *cursor) int {
	if n == nil {
		return 0
	}
	return 1 + depth(n.parent)
}

func sendRemoved(ctx context.Context, from *cursor) (diff Diff, err error) {
	diff = Diff{
		Type: RemovedDiff,
		Key:  from.CurrentKey(),
		From: from.currentValue(),
	}

	if err = from.advance(ctx); err != nil {
		return Diff{}, err
	}
	return
}

func sendAdded(ctx context.Context, to *cursor) (diff Diff, err error) {
	diff = Diff{
		Type: AddedDiff,
		Key:  to.CurrentKey(),
		To:   to.currentValue(),
	}

	if err = to.advance(ctx); err != nil {
		return Diff{}, err
	}
	return
}

func sendModified(ctx context.Context, from, to *cursor) (diff Diff, err error) {
	diff = Diff{
		Type: ModifiedDiff,
		Key:  from.CurrentKey(),
		From: from.currentValue(),
		To:   to.currentValue(),
	}

	if err = from.advance(ctx); err != nil {
		return Diff{}, err
	}
	if err = to.advance(ctx); err != nil {
		return Diff{}, err
	}
	return
}

func skipCommon(ctx context.Context, from, to *cursor) (err error) {
	// track when |from.parent| and |to.parent| change
	// to avoid unnecessary comparisons.
	parentsAreNew := true

	for from.Valid() && to.Valid() {
		if !equalItems(from, to) {
			// found the next difference
			return nil
		}

		if parentsAreNew {
			if equalParents(from, to) {
				// if our parents are equal, we can search for differences
				// faster at the next highest tree Level.
				if err = skipCommonParents(ctx, from, to); err != nil {
					return err
				}
				continue
			}
			parentsAreNew = false
		}

		// if one of the cursors is at the end of its node, it will
		// need to Advance its parent and fetch a new node. In this
		// case we need to Compare parents again.
		parentsAreNew = from.atNodeEnd() || to.atNodeEnd()

		if err = from.advance(ctx); err != nil {
			return err
		}
		if err = to.advance(ctx); err != nil {
			return err
		}
	}

	return err
}

func skipCommonParents(ctx context.Context, from, to *cursor) (err error) {
	err = skipCommon(ctx, from.parent, to.parent)
	if err != nil {
		return err
	}

	if from.parent.Valid() {
		if err = from.fetchNode(ctx); err != nil {
			return err
		}
		from.skipToNodeStart()
	} else {
		from.invalidateAtEnd()
	}

	if to.parent.Valid() {
		if err = to.fetchNode(ctx); err != nil {
			return err
		}
		to.skipToNodeStart()
	} else {
		to.invalidateAtEnd()
	}

	return
}

// todo(andy): assumes equal byte representations
func equalItems(from, to *cursor) bool {
	return bytes.Equal(from.CurrentKey(), to.CurrentKey()) &&
		bytes.Equal(from.currentValue(), to.currentValue())
}

func equalParents(from, to *cursor) (eq bool) {
	if from.parent != nil && to.parent != nil {
		eq = equalItems(from.parent, to.parent)
	}
	return
}

func equalcursorValues(from, to *cursor) bool {
	return bytes.Equal(from.currentValue(), to.currentValue())
}
