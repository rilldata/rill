import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";

export type Param = string;

export type PathOption = {
  label: string;
  depth?: number;
  href?: string;
  preloadData?: boolean;
  section?: string;
  pill?: string;
};

export type PathOptions = {
  options: Map<Param, PathOption>;
  componentOverride?: typeof BreadcrumbItem;
};
