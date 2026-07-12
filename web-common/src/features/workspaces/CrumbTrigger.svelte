<script lang="ts">
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { getIconComponent } from "../entity-management/resource-icon-mapping";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let kind: ResourceKind | undefined;
  export let label: string | undefined;
  export let size = 12;
  export let filePath: string;

  $: icon = getIconComponent(kind, filePath);
</script>

<span class="gap-x-1.5 items-center font-medium flex" title={label}>
  <span class="flex-none">
    <svelte:component this={icon} size="{size}px" />
  </span>
  <span class="truncate">
    {label ?? m.common_loading()}
  </span>
</span>
