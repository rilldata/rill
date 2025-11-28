import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";

export interface ResourceNodeData extends Record<string, unknown> {
  resource: V1Resource;
  kind?: ResourceKind;
  label: string;
  // transient UI flag to emphasize nodes along the traced path
  routeHighlighted?: boolean;
  // true when the node represents the seeded/root resource for the graph card
  isRoot?: boolean;
}
