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
  // When set, a heading with this label is rendered above consecutive items
  // sharing the same group. Used to visually sort dropdown items by tag.
  groupLabel?: string;
  // When set, the item is rendered with a submenu containing these options.
  // The submenu entries link directly via their own `href`.
  subOptions?: Map<Param, PathOption>;
};

export type PathOptions = {
  options: Map<Param, PathOption>;
  carryOverSearchParams?: boolean;
};
