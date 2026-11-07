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
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { onMount, tick } from "svelte";
  import type {
    PivotSidebarSection,
    PivotState,
  } from "web-common/src/features/dashboards/pivot/types.ts";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import { ephemeralMeasureDialog } from "../ephemeral-measures/dialog-store";
  import PivotDrag from "./PivotDrag.svelte";
  import PivotTagRow from "./PivotTagRow.svelte";
  import { timePillActions, timePillSelectors } from "./time-pill-store";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { hasOpenDashboardModal } from "@rilldata/web-common/features/dashboards/shortcuts/dashboard-shortcuts";

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

  const { exploreName, dashboardStore, validSpecStore } = getStateManagers();

  $: ({ rows, columns, tableMode } = pivotState);
  $: splitColumns = splitPivotChips(columns);
  $: ephemeralMeasureNames = new Set(
    $dashboardStore.ephemeralMeasures?.map((def) => def.name) ?? [],
  );

  function editEphemeralMeasure(id: string) {
    const def = $dashboardStore.ephemeralMeasures?.find((d) => d.name === id);
    if (def) ephemeralMeasureDialog.set({ def });
  }

  function deleteEphemeralMeasure(id: string) {
    metricsExplorerStore.removeEphemeralMeasure(
      $exploreName,
      id,
      $validSpecStore.data?.explore,
    );
  }

  let sidebarHeight = 0;
  let searchText = "";
  let selectedTag: string | null = null;
  // Rendered width of the auto-sized tags column; seeds the resizer until the
  // user drags it to an explicit width.
  let tagsColMeasured: number = TAG_COLUMN.pivot.MIN;

  // i18n-ignore: stable drag-zone identifiers compared by DragList
  const TIME_ZONE: PivotSidebarSection = "Time";
  // i18n-ignore: stable drag-zone identifiers compared by DragList
  const MEASURES_ZONE: PivotSidebarSection = "Measures";
  // i18n-ignore: stable drag-zone identifiers compared by DragList
  const DIMENSIONS_ZONE: PivotSidebarSection = "Dimensions";

  onMount(() => {
    timePillActions.initTimeDimension("time", m.dashboard_time());
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
          title: m.dashboard_time(),
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

  let sidebarElement: HTMLDivElement;
  let lastFieldSection: PivotSidebarSection = TIME_ZONE;

  function fieldRows(section?: string) {
    return Array.from(
      sidebarElement?.querySelectorAll<HTMLElement>(
        `[data-pivot-field]${section ? `[data-pivot-section="${section}"]` : ""}`,
      ) ?? [],
    );
  }

  function focusTarget(target: HTMLElement | null | undefined) {
    target?.focus();
    target?.scrollIntoView({ block: "nearest" });
  }

  function focusField(row: HTMLElement | undefined) {
    if (row) {
      const section = row.dataset.pivotSection;
      if (
        section === TIME_ZONE ||
        section === MEASURES_ZONE ||
        section === DIMENSIONS_ZONE
      )
        lastFieldSection = section;
    }
    focusTarget(row?.querySelector<HTMLButtonElement>(".drag-item button"));
  }

  async function handleSidebarKeydown(event: KeyboardEvent) {
    if (
      event.defaultPrevented ||
      event.metaKey ||
      event.ctrlKey ||
      event.altKey
    )
      return;
    const target = event.target;
    if (!(target instanceof HTMLElement)) return;
    const search = sidebarElement.querySelector<HTMLInputElement>(
      'input[role="textbox"]',
    );
    const tags = Array.from(
      sidebarElement.querySelectorAll<HTMLElement>("[data-pivot-tag]"),
    );
    const tag = target.closest<HTMLElement>("[data-pivot-tag]");
    const row = target.closest<HTMLElement>("[data-pivot-field]");
    const action = target.closest<HTMLButtonElement>(".icons button");
    if (target === search) {
      if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
      event.preventDefault();
      event.stopPropagation();
      const fields = fieldRows();
      if (fields.length)
        focusField(event.key === "ArrowDown" ? fields[0] : fields.at(-1));
      else focusTarget(tags[0]);
      return;
    }
    if (!tag && !row) return;
    if (row && !action && target !== row.querySelector(".drag-item button"))
      return;
    if (
      ![
        "ArrowUp",
        "ArrowDown",
        "ArrowLeft",
        "ArrowRight",
        "Tab",
        "Enter",
        " ",
        "Escape",
      ].includes(event.key)
    )
      return;
    event.preventDefault();
    event.stopPropagation();
    if (event.key === "Escape") {
      if (action) focusField(row ?? undefined);
      else focusTarget(search);
    } else if (event.key === "Tab") {
      const sections = [TIME_ZONE, MEASURES_ZONE, DIMENSIONS_ZONE].filter(
        (section) => fieldRows(section).length,
      );
      const sectionIndex = sections.findIndex(
        (section) => section === row?.dataset.pivotSection,
      );
      const nextIndex = tag
        ? event.shiftKey
          ? sections.length - 1
          : 0
        : (sectionIndex + (event.shiftKey ? -1 : 1) + sections.length) %
          sections.length;
      focusField(fieldRows(sections[nextIndex])[0]);
    } else if (event.key === "ArrowUp" || event.key === "ArrowDown") {
      const offset = event.key === "ArrowUp" ? -1 : 1;
      if (tag) {
        const index = tags.indexOf(tag) + offset;
        focusTarget(
          index < 0 ? search : tags[Math.min(index, tags.length - 1)],
        );
      } else {
        const fields = fieldRows();
        const index = fields.indexOf(row!) + offset;
        if (index < 0) focusTarget(search);
        else focusField(fields[Math.min(index, fields.length - 1)]);
      }
    } else if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
      if (action) {
        const actions = Array.from(
          row?.querySelectorAll<HTMLButtonElement>(".icons button") ?? [],
        );
        const index = actions.indexOf(action);
        if (event.key === "ArrowLeft" && index === 0)
          focusField(row ?? undefined);
        else
          focusTarget(
            actions[
              Math.max(
                0,
                Math.min(
                  actions.length - 1,
                  index + (event.key === "ArrowLeft" ? -1 : 1),
                ),
              )
            ],
          );
      } else if (tag && event.key === "ArrowRight")
        focusField(fieldRows(lastFieldSection)[0] ?? fieldRows()[0]);
      else if (row && event.key === "ArrowLeft") focusTarget(tags[0]);
      else if (row && event.key === "ArrowRight")
        focusTarget(row.querySelector<HTMLButtonElement>(".icons button"));
    } else if (tag) {
      tag.click();
    } else if (row) {
      const placements = Array.from(
        row.querySelectorAll<HTMLButtonElement>("[data-pivot-placement]"),
      );
      if (!action && placements.length > 1) {
        focusTarget(placements[0]);
        return;
      }
      const rowsBefore = fieldRows(row.dataset.pivotSection);
      const index = rowsBefore.indexOf(row);
      const nextRow =
        rowsBefore[index + 1] ??
        rowsBefore[index - 1] ??
        fieldRows().find((candidate) => candidate !== row);
      const nextId = nextRow?.dataset.pivotField;
      const activated = action ?? placements[0];
      const hadOpenModal = hasOpenDashboardModal();
      activated?.click();
      await tick();
      if (!hadOpenModal && hasOpenDashboardModal()) return;
      if (row.isConnected) {
        focusField(row);
        return;
      }
      focusField(
        fieldRows().find(
          (candidate) => candidate.dataset.pivotField === nextId,
        ) ?? fieldRows()[0],
      );
    }
  }

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
  bind:this={sidebarElement}
  data-dashboard-picker
  data-dashboard-shortcut-context="pivot-sidebar"
  role="presentation"
  onkeydown={handleSidebarKeydown}
>
  <div class="input-wrapper sticky top-0 z-10 bg-surface-background">
    <Search
      theme
      background
      bind:value={searchText}
      onKeyDown={handleSidebarKeydown}
    />
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
        title={TIME_ZONE}
        label={m.dashboard_time()}
        items={timeGrainOptions}
        {tableMode}
      />
      <PivotDrag
        title={MEASURES_ZONE}
        label={m.dashboard_measures()}
        items={filteredMeasures}
        {ephemeralMeasureNames}
        onEditEphemeralMeasure={editEphemeralMeasure}
        onDeleteEphemeralMeasure={deleteEphemeralMeasure}
      >
        <button
          slot="header-action"
          class="ml-auto text-fg-secondary hover:text-fg-primary"
          type="button"
          aria-label={m.dashboard_pivot_ephemeral_create()}
          title={m.dashboard_pivot_ephemeral_create()}
          onclick={() => ephemeralMeasureDialog.set({})}
        >
          <Add size="14px" />
        </button>
      </PivotDrag>
      <PivotDrag
        title={DIMENSIONS_ZONE}
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
