<script lang="ts">
  import {
    resourceColorMapping,
    resourceIconMapping,
    resourceLabelMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let kind: ResourceKind;

  $: icon = resourceIconMapping[kind];
  $: color = resourceColorMapping[kind];
  $: label = resourceLabelMapping[kind];
</script>

{#if icon && color && label}
  <span
    class="shrink-0 flex items-center gap-x-1 text-[10px] font-medium px-1.5 py-0.5 rounded"
    style="background-color: {color}15; color: {color}"
  >
    <svelte:component this={icon} size={"12px"} {color} />
    {label}
  </span>
{/if}
