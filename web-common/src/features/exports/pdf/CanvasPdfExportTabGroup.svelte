<script lang="ts">
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import type { TabGroup } from "@rilldata/web-common/features/canvas/stores/tab-group";
  import CanvasPdfExportTab from "./CanvasPdfExportTab.svelte";

  // PDF-capture render of a tab group. Unlike the live CanvasTabGroupView
  // (interactive strip, only the active tab mounted), the exported tabs are
  // stacked: each tab contributes a label band naming it followed by its rows,
  // one tab below the previous in strip order. When allTabs is false only the
  // group's active tab is rendered, so the PDF matches what the user sees.
  let {
    group,
    components,
    maxWidth,
    allTabs,
  }: {
    group: TabGroup;
    components: Map<string, BaseCanvasComponent>;
    maxWidth: number;
    allTabs: boolean;
  } = $props();

  const tabs = $derived(group.tabs);
  const activeTabIndex = $derived(group.activeTabIndex);
  const exportTabs = $derived(
    allTabs ? $tabs : $tabs[$activeTabIndex] ? [$tabs[$activeTabIndex]] : [],
  );
</script>

{#each exportTabs as tab (tab.name)}
  <CanvasPdfExportTab groupName={group.name} {tab} {components} {maxWidth} />
{/each}
