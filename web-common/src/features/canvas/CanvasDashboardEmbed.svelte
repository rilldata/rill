<script lang="ts">
  import { type V1Resource } from "@rilldata/web-common/runtime-client";
  import {
    MIN_HEIGHT,
    normalizeSizeArray,
    DEFAULT_DASHBOARD_WIDTH,
  } from "./layout-util";
  import RowWrapper from "./RowWrapper.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ItemWrapper from "./ItemWrapper.svelte";
  import CanvasDashboardWrapper from "./CanvasDashboardWrapper.svelte";
  import { useCanvas } from "./selector";
  export let resource: V1Resource;

  $: ({ instanceId } = $runtime);

  $: meta = resource?.meta;
  $: canvasName = meta?.name?.name as string;

  $: canvas = resource?.canvas;
  $: rows = canvas?.spec?.rows || [];
  $: maxWidth = canvas?.spec?.maxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: canvasResolverQuery = useCanvas(instanceId, canvasName);

  $: canvasData = $canvasResolverQuery?.data;
</script>

{#if canvasName}
  <CanvasDashboardWrapper
    {maxWidth}
    filtersEnabled={canvas?.spec?.filtersEnabled}
    {canvasName}
  >
    {#each rows as { items = [], height = MIN_HEIGHT, heightUnit = "px" }, rowIndex (rowIndex)}
      {@const widths = normalizeSizeArray(items?.map((el) => el?.width ?? 0))}
      {@const types = items?.map(
        ({ component }) =>
          canvasData?.components?.[component ?? ""]?.component?.spec?.renderer,
      )}
      <RowWrapper
        {maxWidth}
        {rowIndex}
        zIndex={50 - rowIndex * 2}
        height="{height}{heightUnit}"
        gridTemplate={widths.map((w) => `${w}fr`).join(" ")}
      >
        {#each items as item, columnIndex (columnIndex)}
          {@const componentResource =
            canvasData?.components?.[item.component ?? ""]}
          <ItemWrapper type={types[columnIndex]} zIndex={4 - columnIndex}>
            <CanvasComponent
              {canvasName}
              {componentResource}
              canvasItem={item}
              id={item.component ?? ""}
            />
          </ItemWrapper>
        {/each}
      </RowWrapper>
    {:else}
      <div class="size-full flex items-center justify-center">
        <p class="text-lg text-gray-500">No components added</p>
      </div>
    {/each}
  </CanvasDashboardWrapper>
{/if}
