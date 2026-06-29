<script lang="ts">
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import EyeIcon from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
  } from "@rilldata/web-common/runtime-client";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { Button } from "../button";
  import Search from "../search/Search.svelte";
  import DashboardMetricsTagRow from "./DashboardMetricsTagRow.svelte";
  import TagFilterBanner from "./TagFilterBanner.svelte";
  import {
    exploreTagColumnWidth,
    TAG_COLUMN,
  } from "@rilldata/web-common/features/dashboards/workspace/dashboard-layout-store";
  import {
    applyHideAllInTag,
    applyOnlyShowTag,
    applyShowAllInTag,
    computeTagVisibility,
    type TagIndex,
  } from "./tag-utils";

  type SelectableItem = MetricsViewSpecMeasure | MetricsViewSpecDimension;

  export let selectedItems: string[];
  export let allItems: SelectableItem[] = [];
  export let tagIndex: TagIndex;
  export let type: "measure" | "dimension" = "measure";
  export let onSelectedChange: (items: string[]) => void;

  let searchText = "";
  let active = false;
  let selectedTag: string | null = null;
  // Rendered width of the auto-sized tags column; seeds the resizer until the
  // user drags it to an explicit width.
  let tagsColMeasured: number = TAG_COLUMN.explore.MIN;

  const toggleButtonBaseClass =
    "flex h-[26px] w-[42px] items-center justify-center rounded-sm text-icon-muted transition-colors hover:bg-surface-hover hover:text-fg-primary active:bg-gray-300 disabled:text-gray-300 disabled:cursor-not-allowed";

  $: allItemsMap = new Map(allItems.map((item) => [item.name, item]));
  $: numAvailable = allItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;
  $: buttonLabel =
    numAvailable === numShown
      ? type === "measure"
        ? m.explore_all_measures()
        : m.explore_all_dimensions()
      : type === "measure"
        ? m.explore_measures_count({
            count: String(numShown),
            total: String(numAvailable),
          })
        : m.explore_dimensions_count({
            count: String(numShown),
            total: String(numAvailable),
          });
  $: tooltipText =
    type === "measure"
      ? m.explore_choose_measures()
      : m.explore_choose_dimensions();

  $: tags = tagIndex.tags;

  $: hasTags = tags.length > 0;
  $: visibleSet = new Set(selectedItems.filter((id) => id));
  $: searchActive = searchText.trim().length > 0;
  $: filterActive = !!selectedTag;
  $: dragEnabled = !searchActive && !filterActive;

  $: filteredTags = searchActive
    ? tags.filter((t) =>
        t.name.toLowerCase().includes(searchText.trim().toLowerCase()),
      )
    : tags;

  // Items visible in the right column, filtered by the selected tag if any.
  $: itemsForRightColumn = filterActive
    ? (tagIndex.itemsByTag.get(selectedTag!) ?? [])
    : allItems;

  // Reset all transient state whenever the popover closes.
  $: if (!active) {
    selectedTag = null;
    searchText = "";
  }

  // If the active tag disappears (e.g. spec changes), drop the filter.
  $: if (selectedTag && !tags.some((t) => t.name === selectedTag)) {
    selectedTag = null;
  }

  function handleSelectedReorder(data: {
    items: Array<{ id: string }>;
    fromIndex: number;
    toIndex: number;
  }) {
    onSelectedChange(data.items.map((item) => item.id));
  }

  function handleHiddenItemClick(data: { item: { id: string } }) {
    onSelectedChange([...selectedItems, data.item.id]);
  }

  function removeSelectedItem(id: string) {
    if (selectedItems.length <= 1) return;
    onSelectedChange(selectedItems.filter((i) => i !== id));
  }

  function showAllItems() {
    onSelectedChange(allItems.map((item) => item.name ?? "").filter(Boolean));
  }

  function hideAllItems() {
    onSelectedChange([selectedItems[0]]);
  }

  function showAllInTag(tagName: string) {
    onSelectedChange(applyShowAllInTag(selectedItems, tagIndex, tagName));
  }

  function hideAllInTag(tagName: string) {
    onSelectedChange(applyHideAllInTag(selectedItems, tagIndex, tagName));
  }

  function showOnlyTag(tagName: string) {
    onSelectedChange(applyOnlyShowTag(selectedItems, tagIndex, tagName));
  }

  function toggleTagFilter(tagName: string) {
    selectedTag = selectedTag === tagName ? null : tagName;
  }

  function clearTagFilter() {
    selectedTag = null;
  }
</script>

<Popover.Root bind:open={active}>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button {...props} type="text" theme label={tooltipText}>
        <div class="flex items-center gap-x-0.5 px-1">
          <strong
            >{buttonLabel}</strong
          >
          <span class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon />
          </span>
        </div>
      </Button>
    {/snippet}
  </Popover.Trigger>
  <Popover.Content
    class={hasTags
      ? "p-0 z-popover text-fg-primary w-[600px]"
      : "p-0 z-popover text-fg-primary"}
    align="start"
    strategy="absolute"
    overflowY="auto"
    overflowX="hidden"
    minHeight="100px"
  >
    <div class="flex flex-col">
      <div class="px-3 pt-3 pb-2 border-b border-border">
        <Search
          bind:value={searchText}
          label={m.explore_search_list()}
          placeholder={hasTags
            ? type === "measure"
              ? m.explore_search_measures_or_tags()
              : m.explore_search_dimensions_or_tags()
            : type === "measure"
              ? m.explore_search_measures()
              : m.explore_search_dimensions()}
          showBorderOnFocus={false}
        />
      </div>

      <div class="flex flex-row" class:divide-x={hasTags}>
        {#if hasTags}
          <!-- Left column: tags. Auto-fits to content (capped) until the user
               drags the divider to an explicit width. -->
          <div
            class="flex flex-col flex-none p-1.5 relative"
            data-testid="tags-section"
            bind:clientWidth={tagsColMeasured}
            style="width: {$exploreTagColumnWidth !== null
              ? `${$exploreTagColumnWidth}px`
              : 'fit-content'}; min-width: {TAG_COLUMN.explore
              .MIN}px; max-width: {$exploreTagColumnWidth !== null
              ? TAG_COLUMN.explore.DRAG_MAX
              : TAG_COLUMN.explore.CAP}px;"
          >
            <Resizer
              direction="EW"
              side="right"
              min={TAG_COLUMN.explore.MIN}
              max={TAG_COLUMN.explore.DRAG_MAX}
              basis={0}
              dimension={$exploreTagColumnWidth ?? tagsColMeasured}
              onUpdate={(d) => exploreTagColumnWidth.set(d === 0 ? null : d)}
            />
            <h3
              class="uppercase font-semibold text-[11px] text-fg-secondary px-2 pt-1 pb-1"
            >
              {m.explore_tags()}
            </h3>
            {#if filteredTags.length === 0}
              <div class="px-2 py-2 text-xs text-fg-secondary">
                {m.explore_no_matching_tags()}
              </div>
            {:else}
              {#each filteredTags as tag (tag.name)}
                <DashboardMetricsTagRow
                  {tag}
                  visibility={computeTagVisibility(
                    tagIndex,
                    visibleSet,
                    tag.name,
                  )}
                  selected={selectedTag === tag.name}
                  onSelect={() => toggleTagFilter(tag.name)}
                  onShowAll={() => showAllInTag(tag.name)}
                  onHideAll={() => hideAllInTag(tag.name)}
                  onShowOnly={() => showOnlyTag(tag.name)}
                />
              {/each}
            {/if}
          </div>
        {/if}

        <!-- Right column: shown/hidden lists. A min width keeps the list usable
             as the tags column is dragged wider within the fixed-width popover. -->
        <div
          class="flex flex-col flex-1 {hasTags ? 'min-w-[240px]' : 'min-w-0'}"
        >
          {#if filterActive && selectedTag}
            <TagFilterBanner tagName={selectedTag} onClear={clearTagFilter} />
          {/if}

          {#key selectedTag}
            {@const shownFiltered = selectedItems
              .filter((id) => id)
              .filter((id) => {
                if (!filterActive) return true;
                const data = allItemsMap.get(id);
                return data?.tags?.includes(selectedTag!) ?? false;
              })
              .map((id) => ({
                id,
                displayName: allItemsMap.get(id)?.displayName ?? id,
              }))}
            {@const hiddenFiltered = itemsForRightColumn
              .map((i) => i.name)
              .filter((id): id is string => !!id && !selectedItems.includes(id))
              .map((id) => ({
                id,
                displayName: allItemsMap.get(id)?.displayName ?? id,
              }))}

            <!-- Shown items -->
            <div class="p-1.5" data-testid="shown-section">
              <DraggableList
                items={shownFiltered}
                bind:searchValue={searchText}
                minHeight="auto"
                maxHeight="300px"
                draggable={dragEnabled}
                onReorder={handleSelectedReorder}
              >
                {#snippet header()}
                  <div
                    class="flex-none flex w-full py-1.5 pb-1 justify-between px-2 sticky top-0 from-popover from-80% to-transparent bg-gradient-to-b z-10"
                  >
                    <h3
                      class="uppercase font-semibold text-[11px] text-fg-secondary"
                    >
                      {type === "measure"
                        ? m.explore_shown_measures()
                        : m.explore_shown_dimensions()}
                    </h3>
                    {#if shownFiltered.length > 1 && !filterActive}
                      <button
                        class="text-theme-500 pointer-events-auto hover:text-theme-600 font-medium text-xs"
                        onclick={hideAllItems}
                      >
                        {m.explore_hide_all()}
                      </button>
                    {/if}
                  </div>
                {/snippet}

                {#snippet empty()}
                  {searchActive && hasTags && filteredTags.length === 0
                    ? type === "measure"
                      ? m.explore_no_measures_or_tags()
                      : m.explore_no_dimensions_or_tags()
                    : filterActive
                      ? type === "measure"
                        ? m.explore_no_measures_from_tag()
                        : m.explore_no_dimensions_from_tag()
                      : searchActive
                        ? type === "measure"
                          ? m.explore_no_matching_measures_shown()
                          : m.explore_no_matching_dimensions_shown()
                        : type === "measure"
                          ? m.explore_no_measures_shown()
                          : m.explore_no_dimensions_shown()}
                {/snippet}

                {#snippet item({ item })}
                  {@const itemData = allItemsMap.get(item.id)}
                  {@const displayName =
                    itemData?.displayName ??
                    (type === "measure"
                      ? m.explore_unknown_measure()
                      : m.explore_unknown_dimension())}
                  <div class="w-full flex gap-x-1 items-center py-1">
                    {#if itemData?.description || selectedItems.length === 1}
                      <Tooltip.Root delayDuration={200}>
                        <Tooltip.Trigger
                          class="w-full flex gap-x-1 items-center"
                        >
                          {#if dragEnabled}
                            <DragHandle
                              size="16px"
                              className="fill-icon pointer-events-none"
                            />
                          {/if}
                          <span
                            class="truncate min-w-0 flex-1 text-left pointer-events-none text-fg-primary"
                          >
                            {displayName}
                          </span>
                          <button
                            class="{toggleButtonBaseClass} ml-auto"
                            onclick={(e) => {
                              e.stopPropagation();
                              removeSelectedItem(item.id);
                            }}
                            onmousedown={(e) => e.stopPropagation()}
                            disabled={selectedItems.length === 1}
                            class:pointer-events-none={selectedItems.length ===
                              1}
                            class:opacity-50={selectedItems.length === 1}
                            aria-label={m.explore_hide_item({ name: displayName })}
                            data-testid="toggle-visibility-button"
                            type="button"
                          >
                            <EyeIcon size="18px" color="currentColor" />
                          </button>
                        </Tooltip.Trigger>
                        <Tooltip.Content
                          side="right"
                          sideOffset={18}
                          class="bg-popover text-fg-primary z-popover"
                        >
                          {#if selectedItems.length === 1}
                            {type === "measure"
                              ? m.explore_must_show_one_measure()
                              : m.explore_must_show_one_dimension()}
                          {:else}
                            {itemData?.description}
                          {/if}
                        </Tooltip.Content>
                      </Tooltip.Root>
                    {:else}
                      {#if dragEnabled}
                        <DragHandle
                          size="16px"
                          className="fill-icon pointer-events-none"
                        />
                      {/if}
                      <span
                        class="truncate min-w-0 flex-1 text-left pointer-events-none"
                      >
                        {displayName}
                      </span>
                      <button
                        class="{toggleButtonBaseClass} ml-auto"
                        onclick={(e) => {
                          e.stopPropagation();
                          removeSelectedItem(item.id);
                        }}
                        onmousedown={(e) => e.stopPropagation()}
                        disabled={selectedItems.length === 1}
                        class:pointer-events-none={selectedItems.length === 1}
                        class:opacity-50={selectedItems.length === 1}
                        aria-label={m.explore_hide_item({ name: displayName })}
                        type="button"
                      >
                        <EyeIcon size="18px" color="currentColor" />
                      </button>
                    {/if}
                  </div>
                {/snippet}
              </DraggableList>
            </div>

            <!-- Hidden items -->
            {#if hiddenFiltered.length > 0}
              <span class="flex-none h-px bg-border w-full"></span>
              <div class="p-1.5">
                <DraggableList
                  items={hiddenFiltered}
                  bind:searchValue={searchText}
                  minHeight="auto"
                  maxHeight="200px"
                  draggable={false}
                >
                  {#snippet header()}
                    <div
                      class="flex-none flex py-1.5 pb-1 justify-between px-2 sticky top-0 from-popover from-80% to-transparent bg-gradient-to-b"
                    >
                      <h3
                        class="uppercase text-[11px] font-semibold text-fg-secondary"
                      >
                        {type === "measure"
                          ? m.explore_hidden_measures()
                          : m.explore_hidden_dimensions()}
                      </h3>
                      {#if !filterActive}
                        <button
                          class="pointer-events-auto text-theme-500 text-xs font-medium hover:text-theme-600"
                          onclick={showAllItems}
                        >
                          {m.explore_show_all()}
                        </button>
                      {:else}
                        <button
                          class="pointer-events-auto text-theme-500 text-xs font-medium hover:text-theme-600"
                          onclick={() => showAllInTag(selectedTag!)}
                        >
                          {m.explore_show_all_in_tag()}
                        </button>
                      {/if}
                    </div>
                  {/snippet}

                  {#snippet empty()}
                    {searchActive
                      ? type === "measure"
                        ? m.explore_no_matching_hidden_measures()
                        : m.explore_no_matching_hidden_dimensions()
                      : type === "measure"
                        ? m.explore_no_hidden_measures()
                        : m.explore_no_hidden_dimensions()}
                  {/snippet}

                  {#snippet item({ item })}
                    {@const itemData = allItemsMap.get(item.id)}
                    {@const displayName =
                      itemData?.displayName ??
                      (type === "measure"
                        ? m.explore_unknown_measure()
                        : m.explore_unknown_dimension())}
                    <div
                      class="w-full flex gap-x-1 justify-between items-center py-1"
                    >
                      {#if itemData?.description}
                        <Tooltip.Root delayDuration={200}>
                          <Tooltip.Trigger
                            class="w-full flex gap-x-1 justify-between items-center"
                          >
                            <span
                              class="truncate min-w-0 flex-1 text-left pointer-events-none"
                            >
                              {displayName}
                            </span>
                            <button
                              class="{toggleButtonBaseClass} ml-auto"
                              onclick={(e) => {
                                e.stopPropagation();
                                handleHiddenItemClick({ item });
                              }}
                              aria-label={m.explore_show_item({ name: displayName })}
                              data-testid="toggle-visibility-button"
                              type="button"
                            >
                              <EyeOffIcon size="18px" color="currentColor" />
                            </button>
                          </Tooltip.Trigger>
                          <Tooltip.Content
                            side="right"
                            sideOffset={18}
                            class="bg-popover text-fg-primary z-popover"
                          >
                            {itemData.description}
                          </Tooltip.Content>
                        </Tooltip.Root>
                      {:else}
                        <span
                          class="truncate min-w-0 flex-1 text-left pointer-events-none"
                        >
                          {displayName}
                        </span>
                        <button
                          class="{toggleButtonBaseClass} ml-auto"
                          onclick={(e) => {
                            e.stopPropagation();
                            handleHiddenItemClick({ item });
                          }}
                          aria-label={m.explore_show_item({ name: displayName })}
                          data-testid="toggle-visibility-button"
                          type="button"
                        >
                          <EyeOffIcon size="18px" color="currentColor" />
                        </button>
                      {/if}
                    </div>
                  {/snippet}
                </DraggableList>
              </div>
            {/if}
          {/key}
        </div>
      </div>

      {#if filterActive}
        <div
          class="px-3 py-1.5 text-xs text-fg-secondary border-t border-border"
        >
          {type === "measure"
            ? m.explore_clear_tag_filter_to_reorder_measures()
            : m.explore_clear_tag_filter_to_reorder_dimensions()}
        </div>
      {:else if searchActive && !filterActive}
        <div
          class="px-3 py-1.5 text-xs text-fg-secondary border-t border-border"
        >
          {type === "measure"
            ? m.explore_clear_search_to_reorder_measures()
            : m.explore_clear_search_to_reorder_dimensions()}
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>
