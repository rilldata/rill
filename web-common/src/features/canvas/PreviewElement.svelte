<script context="module" lang="ts">
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import CanvasComponent from "./CanvasComponent.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let component: V1CanvasItem;
  export let selected: boolean;
  export let instanceId: string;

  $: componentName = component?.component;
  $: inlineComponent = component?.definedInCanvas;
</script>

{#if componentName && !inlineComponent}
  <CanvasComponent
    {instanceId}
    {i}
    {componentName}
    {selected}
    on:change
    on:contextmenu
  />
{:else if componentName}
  <CanvasComponent
    {instanceId}
    {i}
    {componentName}
    {selected}
    on:change
    on:contextmenu
  />
{/if}
