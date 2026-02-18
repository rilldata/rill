import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

export interface Resource {
  name: string;
  kind: ResourceKind | string;
  state?: string;
  error?: string;
  path?: string;
}
