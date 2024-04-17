<script lang="ts">
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import Chart from "./Chart.svelte";
  import { DEFAULT_WIDTH, DEFAULT_RADIUS } from "./constants";
  import Wrapper from "./Wrapper.svelte";
  import Component from "./Component.svelte";

  export let columns = 20;
  export let components: V1DashboardComponent[];
  export let gap = 4;

  let contentRect: DOMRectReadOnly = new DOMRectReadOnly(0, 0, 0, 0);

  $: gridWidth = contentRect.width;
  $: scale = gridWidth / DEFAULT_WIDTH;
  $: gapSize = DEFAULT_WIDTH * (gap / 1000);
  $: gridCell = DEFAULT_WIDTH / columns;
  $: radius = gridCell * DEFAULT_RADIUS;

  $: maxBottom = components.reduce((max, el) => {
    const bottom = Number(el.height) + Number(el.y);
    return Math.max(max, bottom);
  }, 0);
</script>

<Wrapper
  {scale}
  width={DEFAULT_WIDTH}
  height={maxBottom * gridCell}
  bind:contentRect
  color="bg-gray-100"
>
  {#each components as component, i (i)}
    {#if component.chart && component.width && component.height}
      <Component
        embed
        {i}
        {scale}
        {radius}
        padding={gapSize}
        width={Number(component.width) * gridCell}
        height={Number(component.height) * gridCell}
        left={Number(component.x) * gridCell}
        top={Number(component.y) * gridCell}
      >
        <Chart chartName={component.chart} />
      </Component>
    {/if}
  {/each}
</Wrapper>
