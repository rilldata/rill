package blob

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
)

// objectWithPlan has details on download plan for the remote object
// the plan has following details
// - full download or partial download
// - in case of partial download
//   - strategy of download (head or tail)
//   - size of data to download
type objectWithPlan struct {
	obj           *blob.ListObject
	full          bool
	extractOption *extractOption
}

type extractOption struct {
	limitInBytes uint64
	strategy     runtimev1.Source_ExtractPolicy_Strategy
}

// rowPlanner is an interface that creates download plan of a cloud object
type rowPlanner interface {
	// planFile creates download plan of a object
	planFile(item *blob.ListObject) *objectWithPlan
	// done returns true when download limit is breached
	done() bool
}

// plannerWithGlobalLimits implements rowPlanner interface
// the limitInBytes is a combined limit on all files
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
		obj.extractOption = &extractOption{limitInBytes: r.limitInBytes - r.cumsizeInBytes, strategy: r.strategy}
		r.full = true
	}
	r.cumsizeInBytes += uint64(item.Size)
	return obj
}

func (r *plannerWithGlobalLimits) done() bool {
	return r.full
}

// plannerWithPerFileLimits implements rowPlanner interface
// limitInBytes is on individual file
type plannerWithPerFileLimits struct {
	strategy     runtimev1.Source_ExtractPolicy_Strategy
	limitInBytes uint64
}

func (r *plannerWithPerFileLimits) planFile(item *blob.ListObject) *objectWithPlan {
	return &objectWithPlan{
		obj:           item,
		full:          uint64(item.Size) < r.limitInBytes, // if requested more data than size of file
		extractOption: &extractOption{limitInBytes: r.limitInBytes, strategy: r.strategy},
	}
}

func (r *plannerWithPerFileLimits) done() bool {
	return false
}

// plannerWithoutLimits implements rowPlanner interface
// there are no limits
type plannerWithoutLimits struct{}

func (r *plannerWithoutLimits) planFile(item *blob.ListObject) *objectWithPlan {
	return &objectWithPlan{obj: item, full: true}
}

func (r *plannerWithoutLimits) done() bool {
	return false
}
