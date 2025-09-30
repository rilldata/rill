<script lang="ts">
  import { type V1Resource } from "@rilldata/web-common/runtime-client";
  import { DEFAULT_DASHBOARD_WIDTH } from "./layout-util";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let resource: V1Resource;
  export let navigationEnabled: boolean = true;
  export let dynamicHeight: boolean = false;

  $: ({ instanceId } = $runtime);

  $: meta = resource?.meta;
  $: canvasName = meta?.name?.name as string;

  $: canvas = resource?.canvas;
  $: maxWidth = canvas?.spec?.maxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: ({
    canvasEntity: { components, _rows },
  } = getCanvasStore(canvasName, instanceId));

  $: rows = $_rows;
</script>

{#if canvasName}
  <CanvasDashboardWrapper
    {maxWidth}
    {canvasName}
    filtersEnabled={canvas?.spec?.filtersEnabled}
    {dynamicHeight}
  >
    {#each rows as row, rowIndex (rowIndex)}
      <StaticCanvasRow
        {row}
        {rowIndex}
        {components}
        {maxWidth}
        {navigationEnabled}
      />
    {:else}
      <div class="size-full flex items-center justify-center">
        <p class="text-lg text-gray-500">No components added</p>
      </div>
    {/each}
  </CanvasDashboardWrapper>
{/if}
