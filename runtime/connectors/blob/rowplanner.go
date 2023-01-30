package blob

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
)

type rowPlanner interface {
	planFile(item *blob.ListObject) *objectWithPlan
	isFull() bool
}

type plannerWithGlobalLimits struct {
	cumsizeInBytes uint64
	strategy       runtimev1.Source_ExtractPolicy_Strategy
	limitInBytes   uint64
	full           bool
}

func (r *plannerWithGlobalLimits) planFile(item *blob.ListObject) *objectWithPlan {
	obj := &objectWithPlan{obj: item}
	obj.full = true
	if uint64(item.Size)+r.cumsizeInBytes > r.limitInBytes {
		obj.full = false
		obj.size = r.limitInBytes - r.cumsizeInBytes
		obj.stratety = r.strategy
		r.full = true
	}
	return obj
}

func (r *plannerWithGlobalLimits) isFull() bool {
	return r.full
}

type plannerWithPerFileLimits struct {
	strategy     runtimev1.Source_ExtractPolicy_Strategy
	limitInBytes uint64
	full         bool
}

func (r *plannerWithPerFileLimits) planFile(item *blob.ListObject) *objectWithPlan {
	return &objectWithPlan{obj: item, full: false, size: r.limitInBytes, stratety: r.strategy}
}

func (r *plannerWithPerFileLimits) isFull() bool {
	return r.full
}

type plannerWithoutLimits struct{}

func (r *plannerWithoutLimits) planFile(item *blob.ListObject) *objectWithPlan {
	return &objectWithPlan{obj: item, full: true}
}

func (r *plannerWithoutLimits) isFull() bool {
	return false
}

type objectWithPlan struct {
	obj      *blob.ListObject
	full     bool
	size     uint64
	stratety runtimev1.Source_ExtractPolicy_Strategy
}
