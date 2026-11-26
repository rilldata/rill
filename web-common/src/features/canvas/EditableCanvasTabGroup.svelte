<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import CanvasTabStrip from "./CanvasTabStrip.svelte";
  import EditableBlockRegion, {
    type BlockBuilderProps,
  } from "./EditableBlockRegion.svelte";
  import { tabRowsPath } from "./layout-util";
  import type { TabGroup } from "./stores/tab-group";
  import LevelControls from "./controls/LevelControls.svelte";
  import AddControlDialog from "./controls/editor/AddControlDialog.svelte";
  import type { ControlLevelRef } from "./controls/control-edit";
  import type { CanvasEntity } from "./stores/canvas-entity";

  // A tab group in the editor: the shared block region with the tab strip as its header,
  // editing the active tab's rows.
  // svelte-ignore custom_element_props_identifier
  let {
    group,
    canvas,
    onAddTab,
    onRenameTab,
    onDeleteTab,
    onMoveTab,
    onDuplicateTab,
    onSelect,
    onDropOnTab,
    onSelectBlock,
    onBlockMouseDown,
    onDuplicateBlock,
    ...builder
  }: BlockBuilderProps & {
    group: TabGroup;
    canvas: CanvasEntity;
    onAddTab: (blockIndex: number) => void;
    onRenameTab: (blockIndex: number, tabIndex: number, label: string) => void;
    onDeleteTab: (blockIndex: number, tabIndex: number) => void;
    onMoveTab: (
      blockIndex: number,
      tabIndex: number,
      direction: -1 | 1,
    ) => void;
    onDuplicateTab: (blockIndex: number, tabIndex: number) => void;
    // Called with the selected tab's stable name so the parent can sync URL state.
    onSelect?: (tabName: string) => void;
    // Drop a dragged component onto a tab (cross-container move).
    onDropOnTab: (blockIndex: number, tabIndex: number) => void;
  } = $props();

  const blockIndex = $derived(builder.blockIndex);
  const tabs = $derived(group.tabs);
  const activeTabIndex = $derived(group.activeTabIndex);
  const activeTab = $derived($tabs[$activeTabIndex]);

  // A new viewer control goes on a tab, or on the whole group if the author picks it.
  function controlLevels(tabIndex: number): ControlLevelRef[] {
    return [
      {
        kind: "tab",
        path: ["rows", blockIndex, "tabs", tabIndex],
        name: $tabs[tabIndex]?.displayName ?? "",
      },
      { kind: "group", path: ["rows", blockIndex], name: group.name },
    ];
  }
  // The tab whose menu asked for a new control.
  let addingControlTab = $state<number | null>(null);
</script>

<EditableBlockRegion
  {...builder}
  variant="solid"
  grid={activeTab?.grid}
  target={{ rowsPath: tabRowsPath(blockIndex, $activeTabIndex) }}
  idPrefix={`${blockIndex}-${$activeTabIndex}-`}
  zoneScope={`tab:${group.name}:${activeTab?.name ?? $activeTabIndex}`}
  rowKey={activeTab?.name ?? String($activeTabIndex)}
  emptyLabel={m.canvas_add_widget_to_tab()}
  afterLabel={m.canvas_add_widget_below_tabs()}
>
  {#snippet header()}
    <CanvasTabStrip
      {group}
      maxWidth={builder.maxWidth}
      editable
      dragComponent={builder.dragComponent}
      {onSelect}
      onGroupMouseDown={(event) => onBlockMouseDown(blockIndex, event)}
      onDuplicateGroup={() => onDuplicateBlock(blockIndex)}
      onAddTab={() => onAddTab(blockIndex)}
      onRenameTab={(tabIndex, label) =>
        onRenameTab(blockIndex, tabIndex, label)}
      onDeleteTab={(tabIndex) => onDeleteTab(blockIndex, tabIndex)}
      onMoveTab={(tabIndex, direction) =>
        onMoveTab(blockIndex, tabIndex, direction)}
      onDuplicateTab={(tabIndex) => onDuplicateTab(blockIndex, tabIndex)}
      onAddControl={(tabIndex) => (addingControlTab = tabIndex)}
      onSelectGroup={() => onSelectBlock(group.name)}
      onDropOnTab={(tabIndex) => onDropOnTab(blockIndex, tabIndex)}
    />
    <!-- The viewer controls of the group and of its active tab. -->
    <LevelControls
      canvasEntity={canvas}
      scopes={[
        [group.name],
        ...(activeTab ? [[group.name, activeTab.name]] : []),
      ]}
      add={{
        levels: controlLevels($activeTabIndex),
        levelPath: ["rows", blockIndex, "tabs", $activeTabIndex],
      }}
    />
  {/snippet}
</EditableBlockRegion>

{#if addingControlTab !== null}
  <AddControlDialog
    {canvas}
    levels={controlLevels(addingControlTab)}
    levelPath={["rows", blockIndex, "tabs", addingControlTab]}
    onClose={() => (addingControlTab = null)}
  />
{/if}
