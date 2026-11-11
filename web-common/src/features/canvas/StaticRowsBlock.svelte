<script lang="ts">
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import type { RowsBlock } from "./stores/layout-blocks";

  // The rows of a tab or a container outside the editor (view mode, embeds and the PDF
  // export): plain rows with the same grid as the free canvas. The idPrefix namespaces the
  // row ids so the same row index in different blocks stays unique.
  let {
    block,
    components,
    maxWidth,
    idPrefix,
    navigationEnabled = true,
    activeComponentId = null,
    lazy = true,
    tabGroupName,
    tabName,
  }: {
    block: RowsBlock;
    components: Map<string, BaseCanvasComponent>;
    maxWidth: number;
    idPrefix: string;
    navigationEnabled?: boolean;
    activeComponentId?: string | null;
    // When false, components render eagerly without lazy-load (PDF export).
    lazy?: boolean;
    // The tab group and tab of a tab's rows in the export render (see CanvasPdfExportTab).
    tabGroupName?: string;
    tabName?: string;
  } = $props();

  const rows = $derived(block.grid);
</script>

{#each $rows as row, rowIndex (`${block.name}-${rowIndex}`)}
  <StaticCanvasRow
    {row}
    {rowIndex}
    {components}
    {maxWidth}
    {navigationEnabled}
    {activeComponentId}
    {lazy}
    {idPrefix}
    {tabGroupName}
    {tabName}
  />
{/each}
