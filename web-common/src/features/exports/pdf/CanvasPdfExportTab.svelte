<script lang="ts">
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import CanvasPdfExportControls from "./CanvasPdfExportControls.svelte";
  import StaticRowsBlock from "@rilldata/web-common/features/canvas/StaticRowsBlock.svelte";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import type { RowsBlock } from "@rilldata/web-common/features/canvas/stores/layout-blocks";

  // One exported tab: a label band naming the tab (captured as its own block,
  // see captureTargetsIn) followed by the tab's rows. The idPrefix namespaces
  // the row ids so the same rowIndex in different tabs stays unique.
  let {
    canvasEntity,
    groupName,
    tab,
    components,
    maxWidth,
  }: {
    canvasEntity: CanvasEntity;
    groupName: string;
    tab: RowsBlock;
    components: Map<string, BaseCanvasComponent>;
    maxWidth: number;
  } = $props();
</script>

<section
  id={`pdf-tab-label-${groupName}-${tab.name}`}
  class="pdf-tab-label-row w-full h-fit relative"
  style:max-width="{maxWidth}px"
>
  <h2>{tab.displayName}</h2>
</section>

<!-- The viewer controls of the tab group and of this tab, as text. -->
<CanvasPdfExportControls
  {canvasEntity}
  scopes={[[groupName], [groupName, tab.name]]}
  {maxWidth}
/>

<StaticRowsBlock
  block={tab}
  {components}
  {maxWidth}
  navigationEnabled={false}
  lazy={false}
  idPrefix={`${groupName}-${tab.name}-`}
  tabGroupName={groupName}
  tabName={tab.name}
/>

<style lang="postcss">
  /* Mimics the live tab strip: the tab's name underlined in the accent color on
     a hairline spanning the row, so the PDF shows which tab the rows below
     belong to. */
  .pdf-tab-label-row {
    @apply flex items-end border-b border-gray-200 px-2 pt-3;
  }

  h2 {
    @apply -mb-px w-fit border-b-2 border-primary-500 px-1 pb-1;
    @apply text-sm font-medium text-fg-primary;
  }
</style>
