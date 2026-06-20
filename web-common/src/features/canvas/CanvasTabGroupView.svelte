<script lang="ts">
  import CanvasTabStrip from "./CanvasTabStrip.svelte";
  import StaticCanvasRow from "./StaticCanvasRow.svelte";
  import type { TabGroup } from "./stores/tab-group";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";

  export let group: TabGroup;
  export let maxWidth: number;
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

<div class="tab-group-region" style:max-width="{maxWidth}px">
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
  {/if}
</div>

<style lang="postcss">
  /* Visible boundary marking which widgets belong to the tab versus the free canvas. */
  .tab-group-region {
    @apply w-full mx-auto relative flex flex-col items-center;
    @apply rounded-md border border-gray-200 bg-surface-subtle/40 px-3 py-2;
  }
</style>
