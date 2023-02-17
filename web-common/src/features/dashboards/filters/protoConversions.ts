import { Value } from "@bufbuild/protobuf";
import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import {
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "../../../../../proto/gen/rill/runtime/v1/api_pb";
import { DashboardState } from "../../../../../proto/gen/rill/ui/v1/dashboard_pb";

export function toProto(filters: V1MetricsViewFilter): DashboardState {
  const filtersMessage = new MetricsViewFilter({
    include: toFilterCondProto(filters.include),
    exclude: toFilterCondProto(filters.exclude),
  });
  return new DashboardState({
    filters: filtersMessage,
  });
}

export function fromProto(binary: Uint8Array) {
  return MetricsViewFilter.fromBinary(binary);
}

function toFilterCondProto(conds: Array<MetricsViewFilterCond>) {
  return conds.map(
    (include) =>
      new MetricsViewFilter_Cond({
        name: include.name,
        like: include.like,
        in: include.in.map(
          (v) =>
            new Value({
              kind: {
                case: "stringValue",
                value: v as string,
              },
            })
        ),
      })
  );
}
