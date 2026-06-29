<script lang="ts">
  import CanvasTabStrip from "./CanvasTabStrip.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import ItemWrapper from "./ItemWrapper.svelte";
  import RowWrapper from "./RowWrapper.svelte";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import type { TabGroup } from "./stores/tab-group";

  export let group: TabGroup;
  export let maxWidth: number;
  export let zIndex = 1;
  export let components: Map<string, BaseCanvasComponent>;
  export let navigationEnabled: boolean = true;
  export let activeComponentId: string | null = null;
  export let onSelect: ((tabName: string) => void) | undefined = undefined;

  $: tabs = group.tabs;
  $: activeTabIndex = group.activeTabIndex;
  $: activeTab = $tabs[$activeTabIndex];
  // Subscribe to the active tab's rows. Only the active tab is mounted, so inactive
  // tabs never render their components and therefore never issue queries.
  $: activeRows = activeTab?.grid;
</script>

<RowWrapper
  gridTemplate="12fr"
  {zIndex}
  {maxWidth}
  id={`tab-group-row-${group.name}`}
>
  <ItemWrapper fitContent zIndex={0}>
    <div class="tab-group-region">
      <CanvasTabStrip {group} {maxWidth} {onSelect} />

      {#if activeTab && activeRows}
        {#each $activeRows as row, rowIndex (rowIndex)}
          <StaticCanvasRow
            {row}
            {rowIndex}
            {components}
            {maxWidth}
            {navigationEnabled}
            {activeComponentId}
            idPrefix={`${group.name}-${activeTab.name}-`}
          />
        {/each}

        {#if $activeRows.length === 0}
          <div class="flex items-center justify-center py-8">
            <p class="text-fg-secondary">No components added</p>
          </div>
        {/if}
      {/if}
    </div>
  </ItemWrapper>
</RowWrapper>

<style lang="postcss">
  /* Visible boundary marking which widgets belong to the tab versus the free canvas. */
  .tab-group-region {
    @apply pointer-events-auto relative flex w-full min-w-0 flex-col items-center gap-y-2.5;
    @apply rounded-md border border-gray-200 bg-surface-subtle/40 px-3 py-2;
  }
</style>
