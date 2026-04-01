<script lang="ts">
  import {
    resourceIconMapping,
    resourceLabelMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    resourceKindStyleName,
    type ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let kind: ResourceKind;

  $: icon = resourceIconMapping[kind];
  $: label = resourceLabelMapping[kind];
  $: styleName = resourceKindStyleName(kind);
</script>

{#if icon && label}
  <div
    class="flex items-center gap-x-1 text-[10px] font-medium h-full w-full px-3 {styleName}"
  >
    <svelte:component this={icon} size={"12px"} />
    {label}
  </div>
{/if}
