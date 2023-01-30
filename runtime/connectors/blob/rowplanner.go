package blob

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
)

type rowPlanner interface {
	planFile(item *blob.ListObject) *objectWithPlan
	done() bool
}

// plannerWithGlobalLimits adds download limit to all file as per strategy
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
		obj.extractOption = &extractOption{limtiInBytes: r.limitInBytes - r.cumsizeInBytes, strategy: r.strategy}
		r.full = true
	}
	return obj
}

func (r *plannerWithGlobalLimits) done() bool {
	return r.full
}

// plannerWithPerFileLimits adds download limit to every file as per strategy
type plannerWithPerFileLimits struct {
	strategy     runtimev1.Source_ExtractPolicy_Strategy
	limitInBytes uint64
}

func (r *plannerWithPerFileLimits) planFile(item *blob.ListObject) *objectWithPlan {
	return &objectWithPlan{
		obj:           item,
		full:          false,
		extractOption: &extractOption{limtiInBytes: r.limitInBytes, strategy: r.strategy},
	}
}

func (r *plannerWithPerFileLimits) done() bool {
	return false
}

type plannerWithoutLimits struct{}

func (r *plannerWithoutLimits) planFile(item *blob.ListObject) *objectWithPlan {
	return &objectWithPlan{obj: item, full: true}
}

func (r *plannerWithoutLimits) done() bool {
	return false
}

type objectWithPlan struct {
	obj           *blob.ListObject
	full          bool
	extractOption *extractOption
}

type extractOption struct {
	limtiInBytes uint64
	strategy     runtimev1.Source_ExtractPolicy_Strategy
}
