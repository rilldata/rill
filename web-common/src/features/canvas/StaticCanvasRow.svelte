<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import { normalizeSizeArray } from "./layout-util";
  import type { CanvasEntity } from "./stores/canvas-entity";
  import type { Row } from "./stores/row";

  export let row: Row;
  export let zIndex = 1;
  export let maxWidth: number;
  export let rowIndex: number;
  export let components: CanvasEntity["components"];
  export let heightUnit: string = "px";
  export let navigationEnabled: boolean = true;

  $: ({ height, items: _itemIds, widths: itemWidths } = row);

  $: widths = normalizeSizeArray($itemWidths);

  $: itemIds = $_itemIds;

  $: id = `canvas-row-${rowIndex}`;
</script>

<RowWrapper
  {zIndex}
  {maxWidth}
  height={$height}
  {heightUnit}
  {id}
  gridTemplate={widths.map((w) => `${w}fr`).join(" ")}
>
  {#each itemIds as id, columnIndex (columnIndex)}
    {@const component = components.get(id)}
    <ItemWrapper type={component?.type} zIndex={4 - columnIndex}>
      {#if component}
        <CanvasComponent {component} {navigationEnabled} />
      {:else}
        <ComponentError error="No valid component {id} in project" />
      {/if}
    </ItemWrapper>
  {/each}
</RowWrapper>
