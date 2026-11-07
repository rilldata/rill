<script lang="ts">
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import StaticCanvasRow from "@rilldata/web-common/features/canvas/StaticCanvasRow.svelte";
  import type { Tab } from "@rilldata/web-common/features/canvas/stores/tab-group";

  // One exported tab: a label band naming the tab (captured as its own block,
  // see captureTargetsIn) followed by the tab's rows. The idPrefix namespaces
  // the row ids so the same rowIndex in different tabs stays unique.
  let {
    groupName,
    tab,
    components,
    maxWidth,
  }: {
    groupName: string;
    tab: Tab;
    components: Map<string, BaseCanvasComponent>;
    maxWidth: number;
  } = $props();

  const rows = $derived(tab.grid);
</script>

<section
  id={`pdf-tab-label-${groupName}-${tab.name}`}
  class="pdf-tab-label-row w-full h-fit relative"
  style:max-width="{maxWidth}px"
>
  <h2>{tab.displayName}</h2>
</section>

{#each $rows as row, rowIndex (rowIndex)}
  <StaticCanvasRow
    {row}
    {rowIndex}
    {components}
    {maxWidth}
    navigationEnabled={false}
    lazy={false}
    idPrefix={`${groupName}-${tab.name}-`}
  />
{/each}

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
