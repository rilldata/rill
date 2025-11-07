<script lang="ts">
  import { type V1Resource } from "@rilldata/web-common/runtime-client";
  import { DEFAULT_DASHBOARD_WIDTH } from "./layout-util";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { getCanvasStore } from "./state-managers/state-managers";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Spinner from "../entity-management/Spinner.svelte";
  import { EntityStatus } from "../entity-management/types";

  export let resource: V1Resource;
  export let canvasName: string;
  export let navigationEnabled: boolean = true;

  $: ({ instanceId } = $runtime);

  $: canvas = resource?.canvas;
  $: maxWidth = canvas?.spec?.maxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: ({
    canvasEntity: { components, _rows, firstLoad },
  } = getCanvasStore(canvasName, instanceId));

  $: rows = $_rows;
</script>

{#if canvasName}
  <CanvasDashboardWrapper
    {maxWidth}
    {canvasName}
    filtersEnabled={canvas?.spec?.filtersEnabled}
    embedded
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
        {#if $firstLoad}
          <Spinner status={EntityStatus.Running} size="32px" />
        {:else}
          <p class="text-lg text-gray-500">No components added</p>
        {/if}
      </div>
    {/each}
  </CanvasDashboardWrapper>
{/if}
