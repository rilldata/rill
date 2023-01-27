package blob

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/container"
	"gocloud.dev/blob"
)

// planner keeps items as per extract policy
// it adds item in the container which stops consuming files once it reaches file extract policy limits
type planner struct {
	policy *ExtractPolicy
	// rowPlanner adds support for row extract policy
	rowPlanner rowPlanner
	// keeps collection of objects to be downloaded
	// also adds support for file extract policy
	container container.Container[*blobObject]
}

func (s *planner) Add(item *blob.ListObject) bool {
	if s.IsFull() {
		return false
	}

	obj := s.rowPlanner.planFile(item)
	return s.container.Add(obj)
}

func (s *planner) IsFull() bool {
	return s.container.IsFull() || s.rowPlanner.isFull()
}

func (s *planner) Items() []*blobObject {
	return s.container.Items()
}

func newPlanner(policy *ExtractPolicy) (*planner, error) {
	c, err := ContainerForFileStrategy(policy.FilesStrategy, policy.FilesLimit)
	if err != nil {
		return nil, err
	}

	return &planner{policy: policy, container: c, rowPlanner: plannerForRowStrategy(policy)}, nil
}

func ContainerForFileStrategy(strategy runtimev1.Source_ExtractPolicy_Strategy, limit uint64) (container.Container[*blobObject], error) {
	switch strategy {
	case runtimev1.Source_ExtractPolicy_TAIL:
		return container.NewTailContainer(int(limit), func(obj *blobObject) {})
	case runtimev1.Source_ExtractPolicy_HEAD:
		return container.NewBoundedContainer[*blobObject](int(limit))
	default:
		// No option selected
		return container.NewUnboundedContainer[*blobObject]()
	}
}

func plannerForRowStrategy(policy *ExtractPolicy) rowPlanner {
	if policy.RowsStrategy != runtimev1.Source_ExtractPolicy_UNSPECIFIED {
		if policy.FilesStrategy != runtimev1.Source_ExtractPolicy_UNSPECIFIED {
			// file strategy specified row limits are per file
			return &plannerWithPerFileLimits{strategy: policy.RowsStrategy, limitInBytes: policy.RowsLimitBytes}
		}
		// global policy since file strategy is not specified
		return &plannerWithGlobalLimits{strategy: policy.RowsStrategy, limitInBytes: policy.RowsLimitBytes}
	}
	return &plannerWithoutLimits{}
}

type rowPlanner interface {
	planFile(item *blob.ListObject) *blobObject
	isFull() bool
}

type plannerWithGlobalLimits struct {
	cumsizeInBytes uint64
	strategy       runtimev1.Source_ExtractPolicy_Strategy
	limitInBytes   uint64
	full           bool
}

func (r *plannerWithGlobalLimits) planFile(item *blob.ListObject) *blobObject {
	obj := &blobObject{obj: item}
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

func (r *plannerWithPerFileLimits) planFile(item *blob.ListObject) *blobObject {
	return &blobObject{obj: item, full: false, size: r.limitInBytes, stratety: r.strategy}
}

func (r *plannerWithPerFileLimits) isFull() bool {
	return r.full
}

type plannerWithoutLimits struct{}

func (r *plannerWithoutLimits) planFile(item *blob.ListObject) *blobObject {
	return &blobObject{obj: item, full: true}
}

func (r *plannerWithoutLimits) isFull() bool {
	return false
}

type blobObject struct {
	obj      *blob.ListObject
	full     bool
	size     uint64
	stratety runtimev1.Source_ExtractPolicy_Strategy
}
