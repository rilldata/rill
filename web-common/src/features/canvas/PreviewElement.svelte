<script context="module" lang="ts">
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import CanvasComponent from "./CanvasComponent.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let component: V1CanvasItem;

  $: componentName = component?.component;
  $: inlineComponent = component?.definedInCanvas;
</script>

{#if componentName && !inlineComponent}
  <CanvasComponent {i} {componentName} on:change on:contextmenu />
{:else if componentName}
  <CanvasComponent {i} {componentName} on:change on:contextmenu />
{/if}
