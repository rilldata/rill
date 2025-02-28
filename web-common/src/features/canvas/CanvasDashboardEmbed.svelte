<script lang="ts">
  import {
    createQueryServiceResolveCanvas,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
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

  export let resource: V1Resource;

  $: ({ instanceId } = $runtime);

  $: meta = resource?.meta;
  $: canvasName = meta?.name?.name;

  $: canvas = resource?.canvas;
  $: rows = canvas?.spec?.rows || [];
  $: maxWidth = canvas?.spec?.maxWidth || DEFAULT_DASHBOARD_WIDTH;

  $: canvasResolverQuery = createQueryServiceResolveCanvas(
    instanceId,
    canvasName ?? "",
    {},
    { query: { enabled: !!canvasName } },
  );

  $: canvasData = $canvasResolverQuery.data;
</script>

<CanvasDashboardWrapper
  {maxWidth}
  filtersEnabled={canvas?.spec?.filtersEnabled}
>
  {#each rows as { items = [], height = MIN_HEIGHT, heightUnit = "px" }, rowIndex (rowIndex)}
    {@const widths = normalizeSizeArray(items?.map((el) => el?.width ?? 0))}
    {@const types = items?.map(
      ({ component }) =>
        canvasData?.resolvedComponents?.[component ?? ""]?.component?.spec
          ?.renderer,
    )}
    <RowWrapper
      {maxWidth}
      {rowIndex}
      zIndex={50 - rowIndex * 2}
      height="{height}{heightUnit}"
      gridTemplate={widths.map((w) => `${w}fr`).join(" ")}
    >
      {#each items as item, columnIndex (columnIndex)}
        <ItemWrapper type={types[columnIndex]} zIndex={4 - columnIndex}>
          <CanvasComponent canvasItem={item} id={item.component ?? ""} />
        </ItemWrapper>
      {/each}
    </RowWrapper>
  {:else}
    <div class="size-full flex items-center justify-center">
      <p class="text-lg text-gray-500">No components added</p>
    </div>
  {/each}
</CanvasDashboardWrapper>
