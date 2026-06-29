<script lang="ts">
  import TagFilterBanner from "@rilldata/web-common/components/menu/TagFilterBanner.svelte";
  import type { TagIndex } from "@rilldata/web-common/components/menu/tag-utils";
  import {
    pivotTagColumnWidth,
    TAG_COLUMN,
  } from "@rilldata/web-common/features/dashboards/workspace/dashboard-layout-store";
  import { Search } from "@rilldata/web-common/components/search";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import {
    splitPivotChips,
    splitTagItems,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
  import { type TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { onMount } from "svelte";
  import type { PivotState } from "web-common/src/features/dashboards/pivot/types.ts";
  import PivotDrag from "./PivotDrag.svelte";
  import PivotTagRow from "./PivotTagRow.svelte";
  import { timePillActions, timePillSelectors } from "./time-pill-store";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

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
  // Rendered width of the auto-sized tags column; seeds the resizer until the
  // user drags it to an explicit width.
  let tagsColMeasured: number = TAG_COLUMN.pivot.MIN;

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

<div class="sidebar" class:has-tags={hasTags} bind:clientHeight={sidebarHeight}>
  <div class="input-wrapper sticky top-0 z-10 bg-surface-background">
    <Search theme background bind:value={searchText} />
  </div>

  <div class="body">
    {#if hasTags}
      <!-- Tags pane: auto-fits to content (capped) until the user drags the
           divider to an explicit width. -->
      <div
        class="tags-pane"
        bind:clientWidth={tagsColMeasured}
        style="width: {$pivotTagColumnWidth !== null
          ? `${$pivotTagColumnWidth}px`
          : 'fit-content'}; min-width: {TAG_COLUMN.pivot
          .MIN}px; max-width: min({$pivotTagColumnWidth !== null
          ? TAG_COLUMN.pivot.DRAG_MAX
          : TAG_COLUMN.pivot.CAP}px, {TAG_COLUMN.pivot.PCT_CAP}%);"
      >
        <!-- Resizer lives in the pane (not the scroll area) so it stays
             centered on the separator and never induces a scrollbar. -->
        <Resizer
          direction="EW"
          side="right"
          min={TAG_COLUMN.pivot.MIN}
          max={TAG_COLUMN.pivot.DRAG_MAX}
          basis={0}
          dimension={$pivotTagColumnWidth ?? tagsColMeasured}
          onUpdate={(d) => pivotTagColumnWidth.set(d === 0 ? null : d)}
        />
        <div class="tags-scroll">
          <h3 class="column-header">{m.dashboard_tags()}</h3>
          {#if filteredTags.length === 0}
            <p class="text-fg-secondary my-1 px-2 text-xs">
              {m.dashboard_no_matching_tags()}
            </p>
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
      </div>
    {/if}

    <div class="items-column">
      {#if selectedTag}
        <TagFilterBanner tagName={selectedTag} onClear={clearTagFilter} />
      {/if}

      <PivotDrag
        title="Time"
        label={m.dashboard_time()}
        items={timeGrainOptions}
        {tableMode}
      />
      <PivotDrag
        title="Measures"
        label={m.dashboard_measures()}
        items={filteredMeasures}
      />
      <PivotDrag
        title="Dimensions"
        label={m.dashboard_dimensions()}
        items={filteredDimensions}
        {tableMode}
      />
    </div>
  </div>
</div>

<style lang="postcss">
  .sidebar {
    /* Outer width is owned by the resizable wrapper in PivotDisplay. */
    @apply flex flex-col relative overflow-hidden;
    @apply h-full w-full border-r z-0;
    @apply select-none bg-surface-background;
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

  .tags-pane {
    @apply relative flex-none h-full;
  }

  .tags-scroll {
    @apply flex flex-col h-full w-full py-2 px-2 gap-y-0.5;
    /* Vertical scroll only when the list overflows; never horizontal. */
    @apply overflow-y-auto overflow-x-hidden;
  }

  .items-column {
    @apply flex flex-col flex-1 min-w-0;
    @apply overflow-y-auto overflow-x-hidden;
  }

  .column-header {
    @apply uppercase font-semibold text-[10px];
    @apply px-1.5 pt-1 pb-1 text-fg-secondary;
  }
</style>
