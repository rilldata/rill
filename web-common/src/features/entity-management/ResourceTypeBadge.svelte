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
  export let showLabel = true;

  $: icon = resourceIconMapping[kind];
  $: label = resourceLabelMapping[kind];
  $: styleName = resourceKindStyleName(kind);
</script>

{#if icon && label}
  <span
    class="shrink-0 flex items-center {showLabel
      ? 'gap-x-1'
      : ''} text-[10px] font-medium {showLabel
      ? 'px-1.5'
      : 'px-1'} py-0.5 rounded {styleName}"
    title={showLabel ? undefined : label}
  >
    {#if showIcon}
      <svelte:component this={icon} size={"12px"} />
    {/if}
    {#if showLabel}
      {label}
    {/if}
  </span>
{/if}
