import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { Snippet } from "svelte";

export type Param = string;

export type PathOption = {
  label: string;
  depth?: number;
  href?: string;
  // URL segment for this option when it differs from its key in the options map,
  // e.g. options keyed by `kind/name` that link to `/<section>/<name>`.
  param?: string;
  preloadData?: boolean;
  section?: string;
  pill?: string;
  resourceKind?: ResourceKind;
};

export type PathOptions = {
  options: Map<Param, PathOption>;
  // Key of the current page's option, when the map is not keyed by the URL segment.
  // Defaults to the lowercased URL segment.
  currentId?: Param;
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
  // Key of the current page's option in `options`.
  current: Param;
  currentPath: (string | undefined)[];
  depth: number;
  onSelect: ((id: string) => void) | undefined;
  linkMaker: LinkMaker;
};
