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
  groupId?: string;
  groupLabel?: string;
  groupOrder?: number;
};

export type PathOptions = Map<Param, PathOption>;

export type PathOptionEntry = [Param, PathOption];

export type PathOptionGroup = {
  id: string;
  label: string;
  options: PathOptionEntry[];
};

export type PathOptionsWithGroups = PathOptions & {
  groups?: PathOptionGroup[];
};
