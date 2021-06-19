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

package datas

import (
	"context"

	"github.com/dolthub/dolt/go/store/nomdl"
	"github.com/dolthub/dolt/go/store/types"
)

const (
	WorkingSetName      = "WorkingSet"
	WorkspaceMetaField  = "meta"
	WorkingRootRefField = "workingRootRef"
	StagedRootRefField  = "stagedRootRef"
	MergeStateField     = "mergeState"
)

const (
	MergeStateName      = "MergeState"
	MergeStateCommitField = "commit"
	MergeStateWorkingPreMergeField = "working_pre_merge"
)

type WorkingSetMeta struct {
	Meta types.Struct
}

var workingSetTemplate = types.MakeStructTemplate(WorkingSetName, []string{MergeStateField, StagedRootRefField, WorkingRootRefField})
var valueWorkingSetType = nomdl.MustParseType(`Struct WorkingSet {
				mergeState?: Ref<Value>,
				stagedRootRef?:  Ref<Value>,
        workingRootRef:  Ref<Value>,
}`)

var mergeStateTemplate = types.MakeStructTemplate(MergeStateName, []string{MergeStateCommitField, MergeStateWorkingPreMergeField})
var valueMergeStateType = nomdl.MustParseType(`Struct MergeState {
        commit:  Ref<Value>,
				workingPreMerge:  Ref<Value>,
}`)

type MergeState struct {
	Commit          string `json:"commit"`
	PreMergeWorking string `json:"workingPreMerge"`
}

// NewWorkingSet creates a new working set object.
// A working set is a value that has been persisted but is not necessarily referenced by a Commit. As the name implies,
// it's storage for data changes that have not yet been incorporated into the commit graph but need durable storage.
//
// A working set struct has the following type:
//
// ```
// struct WorkingSet {
//   meta: M,
//   workingRootRef: R,
//   stagedRootRef: R,
//   mergeState: M,
// }
// ```
// where M is a struct type and R is a ref type.
func NewWorkingSet(_ context.Context, workingRef types.Ref, stagedRef, mergeStateRef *types.Ref) (types.Struct, error) {
	fields := make(types.StructData)
	fields[WorkingRootRefField] = workingRef

	if stagedRef != nil {
		fields[StagedRootRefField] = stagedRef
	}
	if mergeStateRef != nil {
		fields[MergeStateField] = mergeStateRef
	}

	return types.NewStruct(workingRef.Format(), WorkingSetName, fields)
}

func IsWorkingSet(v types.Value) (bool, error) {
	if s, ok := v.(types.Struct); !ok {
		return false, nil
	} else {
		return types.IsValueSubtypeOf(s.Format(), v, valueWorkingSetType)
	}
}
