<script lang="ts">
  import TagFilterBanner from "@rilldata/web-common/components/menu/TagFilterBanner.svelte";
  import type { TagIndex } from "@rilldata/web-common/components/menu/tag-utils";
  import { Search } from "@rilldata/web-common/components/search";
  import {
    splitPivotChips,
    splitTagItems,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { type TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import type { PivotState } from "web-common/src/features/dashboards/pivot/types.ts";
  import { dragDataStore } from "./DragList.svelte";
  import PivotDrag from "./PivotDrag.svelte";
  import PivotPortalItem from "./PivotPortalItem.svelte";
  import PivotTagRow from "./PivotTagRow.svelte";
  import { timePillActions, timePillSelectors } from "./time-pill-store";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";

  export let pivotState: PivotState;
  export let measures: PivotChipData[];
  export let dimensions: PivotChipData[];
  export let combinedTagIndex: TagIndex;
  export let dimensionTagIndex: TagIndex;
  export let measureTagIndex: TagIndex;
  export let exploreName: string;
  export let timeControlsForPillActions: Pick<
    TimeControlState,
    "timeStart" | "timeEnd" | "minTimeGrain"
  >;

  $: ({ rows, columns, tableMode } = pivotState);
  $: splitColumns = splitPivotChips(columns);

  let sidebarHeight = 0;
  let searchText = "";
  let selectedTag: string | null = null;

  // Tag drag-and-drop state. Mirrors the chip-drag flow in DragList.svelte
  // but coordinates from the sidebar level since tag rows are not chips.
  const TAG_DRAG_THRESHOLD_PX = 4;
  let tagPendingDrag: {
    tagName: string;
    dimensions: PivotChipData[];
    measures: PivotChipData[];
    rect: DOMRect;
    startX: number;
    startY: number;
    offsetX: number;
    offsetY: number;
  } | null = null;
  let tagDragActive = false;
  let tagDragPosition = { left: 0, top: 0 };
  let tagDragOffset = { x: 0, y: 0 };
  let tagDragChip: PivotChipData | null = null;

  onMount(() => {
    timePillActions.initTimeDimension("time", "Time");
  });

  $: if (
    timeControlsForPillActions.timeStart &&
    timeControlsForPillActions.timeEnd
  ) {
    timePillActions.setTimeControls(
      timeControlsForPillActions.timeStart,
      timeControlsForPillActions.timeEnd,
      timeControlsForPillActions.minTimeGrain,
    );
  }

  $: if (rows && columns) {
    timePillActions.updateUsedGrains("time", rows, splitColumns.dimension);
  }

  $: shouldShowTimePill = timePillSelectors.getAllGrainsUsed("time");

  $: timeGrainOptions = !$shouldShowTimePill
    ? [
        {
          id: "time",
          title: "Time",
          type: PivotChipType.Time,
        },
      ]
    : [];

  $: tags = combinedTagIndex.tags;
  $: hasTags = tags.length > 0;

  // Drop the selection if the active tag disappears from the spec
  // (e.g. the user edits the metrics view).
  $: if (selectedTag && !tags.some((t) => t.name === selectedTag)) {
    selectedTag = null;
  }

  $: namesInSelectedTag = selectedTag
    ? new Set(
        (combinedTagIndex.itemsByTag.get(selectedTag) ?? [])
          .map((item) => item.name)
          .filter((n): n is string => !!n),
      )
    : null;

  $: filteredTags = searchText.trim()
    ? tags.filter((t) =>
        t.name.toLowerCase().includes(searchText.trim().toLowerCase()),
      )
    : tags;

  $: filteredMeasures = filterItems(measures, searchText, namesInSelectedTag);
  $: filteredDimensions = filterItems(
    dimensions,
    searchText,
    namesInSelectedTag,
  );

  function filterItems(
    fullList: PivotChipData[],
    search: string,
    tagNames: Set<string> | null,
  ) {
    const lowerSearch = search.trim().toLowerCase();
    return fullList.filter((chip) => {
      if (tagNames && !tagNames.has(chip.id)) return false;
      if (lowerSearch && !chip.title.toLowerCase().includes(lowerSearch))
        return false;
      return true;
    });
  }

  function toggleTagFilter(tagName: string) {
    selectedTag = selectedTag === tagName ? null : tagName;
  }

  function clearTagFilter() {
    selectedTag = null;
  }

  function tagItemsFor(tagName: string) {
    return splitTagItems(tagName, dimensionTagIndex, measureTagIndex);
  }

  function handleTagDragStart(
    e: MouseEvent,
    tagName: string,
    items: { dimensions: PivotChipData[]; measures: PivotChipData[] },
    rect: DOMRect,
  ) {
    tagPendingDrag = {
      tagName,
      dimensions: items.dimensions,
      measures: items.measures,
      rect,
      startX: e.clientX,
      startY: e.clientY,
      offsetX: e.clientX - rect.left,
      offsetY: e.clientY - rect.top,
    };
    window.addEventListener("mousemove", detectTagDragStart);
    window.addEventListener("mouseup", handleTagGlobalMouseUp, { once: true });
  }

  function detectTagDragStart(e: MouseEvent) {
    if (!tagPendingDrag || tagDragActive) return;
    const moved =
      Math.abs(e.clientX - tagPendingDrag.startX) >= TAG_DRAG_THRESHOLD_PX ||
      Math.abs(e.clientY - tagPendingDrag.startY) >= TAG_DRAG_THRESHOLD_PX;
    if (!moved) return;
    beginTagDrag();
  }

  function beginTagDrag() {
    if (!tagPendingDrag) return;
    tagDragActive = true;
    window.removeEventListener("mousemove", detectTagDragStart);

    const { tagName, dimensions, measures, rect, offsetX, offsetY } =
      tagPendingDrag;

    tagDragPosition = { left: rect.left, top: rect.top };
    tagDragOffset = { x: offsetX, y: offsetY };

    // Synthetic chip used by PivotPortalItem to render the floating preview.
    // Pure-measure tags render with the (rectangular) measure chip styling;
    // mixed and pure-dimension tags render with the (rounded) dimension shape.
    const chipType =
      dimensions.length === 0 && measures.length > 0
        ? PivotChipType.Measure
        : PivotChipType.Dimension;
    tagDragChip = {
      id: `__tag__:${tagName}`,
      title: tagName,
      type: chipType,
    };

    dragDataStore.set({
      source: "tags",
      width: rect.width,
      chip: tagDragChip,
      tagPayload: { tagName, dimensions, measures },
    });
  }

  function handleTagGlobalMouseUp() {
    window.removeEventListener("mousemove", detectTagDragStart);
    if (!tagDragActive) {
      // Mousedown without movement: treat as a click on the tag row.
      // Toggles the filter selection. This avoids needing a separate
      // onclick handler that would race with the drag setup.
      if (tagPendingDrag) {
        toggleTagFilter(tagPendingDrag.tagName);
      }
      tagPendingDrag = null;
      return;
    }
    resetTagDrag();
  }

  function resetTagDrag() {
    tagDragActive = false;
    tagDragChip = null;
    tagPendingDrag = null;
    dragDataStore.set(null);
    window.removeEventListener("mousemove", detectTagDragStart);
  }

  function addTagToRows(tagName: string, replace: boolean) {
    const { dimensions: dims } = tagItemsFor(tagName);
    if (replace) {
      metricsExplorerStore.replacePivotRows(exploreName, dims);
      return;
    }
    if (dims.length === 0) return;
    metricsExplorerStore.addPivotFields(exploreName, dims, "rows");
  }

  function addTagToColumns(tagName: string, replace: boolean) {
    const { dimensions: dims, measures: meas } = tagItemsFor(tagName);
    const all = [...dims, ...meas];
    if (replace) {
      metricsExplorerStore.replacePivotColumns(exploreName, all);
      return;
    }
    if (all.length === 0) return;
    metricsExplorerStore.addPivotFields(exploreName, all, "columns");
  }

  function autoArrangeTag(tagName: string, replace: boolean) {
    const { dimensions: dims, measures: meas } = tagItemsFor(tagName);
    if (replace) {
      metricsExplorerStore.replacePivotRows(exploreName, dims);
      metricsExplorerStore.replacePivotColumns(exploreName, meas);
      return;
    }
    if (dims.length > 0) {
      metricsExplorerStore.addPivotFields(exploreName, dims, "rows");
    }
    if (meas.length > 0) {
      metricsExplorerStore.addPivotFields(exploreName, meas, "columns");
    }
  }
</script>

<div
  class="sidebar"
  class:has-tags={hasTags}
  bind:clientHeight={sidebarHeight}
  transition:slide={{ axis: "x" }}
>
  <div class="input-wrapper sticky top-0 z-10 bg-surface-background">
    <Search theme background bind:value={searchText} />
  </div>

  <div class="body">
    {#if hasTags}
      <div class="tags-column">
        <h3 class="column-header">Tags</h3>
        {#if filteredTags.length === 0}
          <p class="text-fg-secondary my-1 px-2 text-xs">No matching tags</p>
        {:else}
          {#each filteredTags as tag (tag.name)}
            {@const items = tagItemsFor(tag.name)}
            <PivotTagRow
              {tag}
              dimensions={items.dimensions}
              measures={items.measures}
              selected={selectedTag === tag.name}
              onAddRows={(replace) => addTagToRows(tag.name, replace)}
              onAddColumns={(replace) => addTagToColumns(tag.name, replace)}
              onAutoArrange={(replace) => autoArrangeTag(tag.name, replace)}
              onDragStart={(e, rect) =>
                handleTagDragStart(e, tag.name, items, rect)}
            />
          {/each}
        {/if}
      </div>
    {/if}

    <div class="items-column">
      {#if selectedTag}
        <TagFilterBanner tagName={selectedTag} onClear={clearTagFilter} />
      {/if}

      <PivotDrag title="Time" items={timeGrainOptions} {tableMode} />
      <PivotDrag title="Measures" items={filteredMeasures} />
      <PivotDrag title="Dimensions" items={filteredDimensions} {tableMode} />
    </div>
  </div>
</div>

{#if tagDragActive && tagDragChip}
  <PivotPortalItem
    item={tagDragChip}
    offset={tagDragOffset}
    position={tagDragPosition}
    removable={false}
    onRelease={resetTagDrag}
  />
{/if}

<style lang="postcss">
  .sidebar {
    @apply flex flex-col relative overflow-hidden;
    @apply h-full border-r z-0 w-60;
    transition-property: width;
    will-change: width;
    @apply select-none bg-surface-background;
  }

  .sidebar.has-tags {
    width: 400px;
  }

  .input-wrapper {
    @apply flex w-full h-fit items-center;
    @apply border-b;
    @apply gap-x-2 p-2;
  }

  .body {
    @apply flex flex-row flex-1 min-h-0;
  }

  .has-tags .body {
    @apply divide-x;
  }

  .tags-column {
    @apply flex flex-col flex-none w-40 py-2 px-2;
    @apply overflow-y-auto gap-y-0.5;
  }

  .items-column {
    @apply flex flex-col flex-1 min-w-0;
    @apply overflow-y-auto;
  }

  .column-header {
    @apply uppercase font-semibold text-[10px];
    @apply px-1.5 pt-1 pb-1 text-fg-secondary;
  }

</style>
