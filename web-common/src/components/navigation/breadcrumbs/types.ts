import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { Snippet } from "svelte";

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
  content?: Snippet<[BreadcrumbItemDropdownProps]>;
};

export type LinkMaker = (
  current: (string | undefined)[],
  depth: number,
  id: string,
  option: PathOption,
  route: string,
) => string | undefined;

export type BreadcrumbItemDropdownProps = {
  options: Map<Param, PathOption>;
  current: string;
  currentPath: (string | undefined)[];
  depth: number;
  onSelect: ((id: string) => void) | undefined;
  linkMaker: LinkMaker;
};
