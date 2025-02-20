<script context="module" lang="ts">
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import CanvasComponent from "./CanvasComponent.svelte";
</script>

<script lang="ts">
  export let component: V1CanvasItem;

  $: componentName = component?.component;
</script>

{#if componentName}
  <CanvasComponent {componentName} on:change on:contextmenu />
{/if}
