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
  export let showIcon = true;

  $: icon = resourceIconMapping[kind];
  $: label = resourceLabelMapping[kind];
  $: styleName = resourceKindStyleName(kind);
</script>

{#if icon && label}
  <span
    class="shrink-0 flex items-center gap-x-1 text-[10px] font-medium px-1.5 py-0.5 rounded {styleName}"
  >
    {#if showIcon}
      <svelte:component this={icon} size={"12px"} />
    {/if}
    {label}
  </span>
{/if}
