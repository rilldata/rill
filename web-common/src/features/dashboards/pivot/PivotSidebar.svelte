<script lang="ts">
  import TagFilterBanner from "@rilldata/web-common/components/menu/TagFilterBanner.svelte";
  import type { TagIndex } from "@rilldata/web-common/components/menu/tag-utils";
  import { Search } from "@rilldata/web-common/components/search";
  import {
    splitPivotChips,
    splitTagItems,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import { type TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import type { PivotState } from "web-common/src/features/dashboards/pivot/types.ts";
  import PivotDrag from "./PivotDrag.svelte";
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
  export let setRows: (items: PivotChipData[]) => void;
  export let setColumns: (items: PivotChipData[]) => void;
  export let timeControlsForPillActions: Pick<
    TimeControlState,
    "timeStart" | "timeEnd" | "minTimeGrain"
  >;

  $: ({ rows, columns, tableMode } = pivotState);
  $: splitColumns = splitPivotChips(columns);

  let sidebarHeight = 0;
  let searchText = "";
  let selectedTag: string | null = null;

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
              {rows}
              {columns}
              selected={selectedTag === tag.name}
              {setRows}
              {setColumns}
              onSelect={() => toggleTagFilter(tag.name)}
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
