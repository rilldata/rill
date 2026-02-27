import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

export type Param = string;

export type PathOption = {
  label: string;
  depth?: number;
  href?: string;
  preloadData?: boolean;
  section?: string;
  pill?: string;
  resourceKind?: ResourceKind;
};

export type PathOptions = {
  options: Map<Param, PathOption>;
  carryOverSearchParams?: boolean;
};
