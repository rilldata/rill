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

export type PathOptionEntry = {
  id: string;
  option: PathOption;
};

export type PathOptionGroup = {
  name: string;
  label: string;
  items: PathOptionEntry[];
};

export type PathOptions = {
  options: Map<Param, PathOption>;
  groups?: PathOptionGroup[];
  carryOverSearchParams?: boolean;
};
