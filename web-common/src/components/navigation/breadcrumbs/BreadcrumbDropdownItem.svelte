<script lang="ts">
  import { page } from "$app/state";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping.ts";
  import type {
    BreadcrumbItemDropdownProps,
    PathOption,
  } from "@rilldata/web-common/components/navigation/breadcrumbs/types.ts";

  let {
    id,
    option,
    current,
    currentPath = [],
    depth = 0,
    onSelect = undefined,
    linkMaker,
  }: {
    id: string;
    option: PathOption;
  } & Omit<BreadcrumbItemDropdownProps, "options"> = $props();

  let isSelected = $derived(id === current.toLowerCase());
  let Icon = $derived(
    option.resourceKind ? resourceIconMapping[option.resourceKind] : undefined,
  );
</script>

<DropdownMenu.CheckboxItem
  class="cursor-pointer"
  checked={isSelected}
  checkSize={"h-3 w-3"}
  href={linkMaker(currentPath, depth, id, option, page.route.id ?? "")}
  preloadData={option.preloadData}
  onclick={() => {
    if (onSelect) {
      onSelect(id);
    }
  }}
>
  <span class="text-xs text-fg-secondary flex-grow flex items-center gap-x-1.5">
    {#if Icon}
      <Icon size="12px" />
    {/if}
    {option.label}
  </span>
</DropdownMenu.CheckboxItem>
