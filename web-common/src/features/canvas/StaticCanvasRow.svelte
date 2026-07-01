<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import CanvasComponent from "./CanvasComponent.svelte";
  import ItemWrapper from "./ItemWrapper.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import { normalizeSizeArray } from "./layout-util";
  import type { Row } from "./stores/row";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";

  export let row: Row;
  export let zIndex = 1;
  export let maxWidth: number;
  export let rowIndex: number;
  export let components: Map<string, BaseCanvasComponent>;
  export let heightUnit: string = "px";
  export let navigationEnabled: boolean = true;
  export let activeComponentId: string | null = null;
  export let idPrefix: string = "";

  $: ({ height, items: _itemIds, widths: itemWidths } = row);

  $: widths = normalizeSizeArray($itemWidths);

  $: itemIds = $_itemIds;

  $: id = `canvas-row-${idPrefix}${rowIndex}`;
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
        <CanvasComponent
          {component}
          {navigationEnabled}
          active={activeComponentId === id}
        />
      {:else}
        <ComponentError error={m.canvas_no_valid_component({ id })} />
      {/if}
    </ItemWrapper>
  {/each}
</RowWrapper>
