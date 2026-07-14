<script lang="ts">
  import PanButton from "./PanButton.svelte";
  import type { PlotBounds } from "./types";

  export let plotBounds: PlotBounds;
  export let canPanLeft: boolean;
  export let canPanRight: boolean;
  export let onPanLeft: (() => void) | undefined = undefined;
  export let onPanRight: (() => void) | undefined = undefined;

  $: midY = plotBounds.top + plotBounds.height / 2;
  $: leftX = plotBounds.left - 20;
  $: rightX = plotBounds.left + plotBounds.width - 14;
</script>

{#if canPanLeft && onPanLeft}
  <g transform="translate({leftX}, {midY})">
    <PanButton onClick={onPanLeft} />
  </g>
{/if}
{#if canPanRight && onPanRight}
  <g transform="translate({rightX}, {midY})">
    <PanButton onClick={onPanRight} flip />
  </g>
{/if}
