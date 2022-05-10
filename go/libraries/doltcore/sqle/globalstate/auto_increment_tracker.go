// Copyright 2021 Dolthub, Inc.
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

package globalstate

import (
	"context"
	"math"
	"sync"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
)

// CoerceAutoIncrementValue converts |val| into an AUTO_INCREMENT sequence value
func CoerceAutoIncrementValue(val interface{}) (uint64, error) {
	switch typ := val.(type) {
	case float32:
		val = math.Round(float64(typ))
	case float64:
		val = math.Round(typ)
	}

	var err error
	val, err = sql.Uint64.Convert(val)
	if err != nil {
		return 0, err
	}
	if val == nil || val == uint64(0) {
		return 0, nil
	}
	return val.(uint64), nil
}

// AutoIncrementTracker knows how to track the latest auto-increment values for every table across all sessions.
type AutoIncrementTracker interface {
	// Current returns the current AUTO_INCREMENT value for |tableName|.
	Current(tableName string) uint64
	// Next returns the next AUTO_INCREMENT value for |tableName|, considering the provided |insertVal|.
	Next(tbl string, insertVal interface{}) (uint64, error)
	// Set sets the current AUTO_INCREMENT value for |tableName|.
	Set(tableName string, val uint64)
	// AddNewTable adds |tablename| to the AutoIncrementTracker.
	AddNewTable(tableName string)
	// DropTable drops |tablename| from the AutoIncrementTracker.
	// TODO: this doesn't respect transaction boundaries
	DropTable(tableName string)
}

func NewAutoIncrementTracker(ctx context.Context, ws *doltdb.WorkingSet) (AutoIncrementTracker, error) {
	ait := autoIncrementTracker{
		wsRef:     ws.Ref(),
		sequences: make(map[string]uint64),
		mu:        &sync.Mutex{},
	}

	// collect auto increment values
	err := ws.WorkingRoot().IterTables(ctx, func(name string, table *doltdb.Table, sch schema.Schema) (bool, error) {
		ok := schema.HasAutoIncrement(sch)
		if !ok {
			return false, nil
		}
		seq, err := table.GetAutoIncrementValue(ctx)
		if err != nil {
			return true, err
		}
		ait.sequences[name] = seq
		return false, nil
	})

	return ait, err
}

type autoIncrementTracker struct {
	wsRef     ref.WorkingSetRef
	sequences map[string]uint64
	mu        *sync.Mutex
}

func (a autoIncrementTracker) Current(tableName string) uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sequences[tableName]
}

func (a autoIncrementTracker) Next(tbl string, insertVal interface{}) (uint64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	given, err := CoerceAutoIncrementValue(insertVal)
	if err != nil {
		return 0, err
	}

	curr := a.sequences[tbl]

	if given == 0 {
		// |given| is 0 or NULL
		a.sequences[tbl]++
		return curr, nil
	}

	if given >= curr {
		a.sequences[tbl] = given
		a.sequences[tbl]++
		return given, nil
	}

	// |given| < curr
	return given, nil
}

func (a autoIncrementTracker) Set(tableName string, val uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sequences[tableName] = val
}

func (a autoIncrementTracker) AddNewTable(tableName string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sequences[tableName] = uint64(1)
}

func (a autoIncrementTracker) DropTable(tableName string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sequences, tableName)
}
