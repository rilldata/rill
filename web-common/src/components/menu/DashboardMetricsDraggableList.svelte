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
  import { Button } from "../button";
  import Search from "../search/Search.svelte";
  import DashboardMetricsTagRow from "./DashboardMetricsTagRow.svelte";
  import TagFilterBanner from "./TagFilterBanner.svelte";
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

  const toggleButtonBaseClass =
    "flex h-[26px] w-[42px] items-center justify-center rounded-sm text-icon-muted transition-colors hover:bg-surface-hover hover:text-fg-primary active:bg-gray-300 disabled:text-gray-300 disabled:cursor-not-allowed";

  $: allItemsMap = new Map(allItems.map((item) => [item.name, item]));
  $: numAvailable = allItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;
  $: numShownString =
    numAvailable === numShown ? "All" : `${numShown} of ${numAvailable}`;
  $: tooltipText = `Choose ${type === "measure" ? "measures" : "dimensions"} to display`;
  $: pluralLabel = type === "measure" ? "measures" : "dimensions";

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
            >{`${numShownString} ${type === "measure" ? "Measures" : "Dimensions"}`}</strong
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
          label="Search list"
          placeholder={hasTags
            ? `Search ${pluralLabel} or tags`
            : `Search ${pluralLabel}`}
          showBorderOnFocus={false}
        />
      </div>

      <div class="flex flex-row" class:divide-x={hasTags}>
        {#if hasTags}
          <!-- Left column: tags -->
          <div
            class="flex flex-col flex-none w-[240px] p-1.5"
            data-testid="tags-section"
          >
            <h3
              class="uppercase font-semibold text-[11px] text-fg-secondary px-2 pt-1 pb-1"
            >
              Tags
            </h3>
            {#if filteredTags.length === 0}
              <div class="px-2 py-2 text-xs text-fg-secondary">
                No matching tags
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

        <!-- Right column: shown/hidden lists -->
        <div class="flex flex-col flex-1 min-w-0">
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
                      Shown {pluralLabel}
                    </h3>
                    {#if shownFiltered.length > 1 && !filterActive}
                      <button
                        class="text-theme-500 pointer-events-auto hover:text-theme-600 font-medium text-xs"
                        onclick={hideAllItems}
                      >
                        Hide all
                      </button>
                    {/if}
                  </div>
                {/snippet}

                {#snippet empty()}
                  {searchActive && hasTags && filteredTags.length === 0
                    ? `No ${pluralLabel} or tags found`
                    : filterActive
                      ? `No ${pluralLabel} from this tag are shown`
                      : searchActive
                        ? `No matching ${pluralLabel} shown`
                        : `No ${pluralLabel} shown`}
                {/snippet}

                {#snippet item({ item })}
                  {@const itemData = allItemsMap.get(item.id)}
                  {@const displayName =
                    itemData?.displayName ??
                    `Unknown ${type === "measure" ? "measure" : "dimension"}`}
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
                            aria-label="Hide {displayName}"
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
                            Must show at least one {type === "measure"
                              ? "measure"
                              : "dimension"}
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
                        aria-label={`Hide ${displayName}`}
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
                        Hidden {pluralLabel}
                      </h3>
                      {#if !filterActive}
                        <button
                          class="pointer-events-auto text-theme-500 text-xs font-medium hover:text-theme-600"
                          onclick={showAllItems}
                        >
                          Show all
                        </button>
                      {:else}
                        <button
                          class="pointer-events-auto text-theme-500 text-xs font-medium hover:text-theme-600"
                          onclick={() => showAllInTag(selectedTag!)}
                        >
                          Show all in tag
                        </button>
                      {/if}
                    </div>
                  {/snippet}

                  {#snippet empty()}
                    {searchActive
                      ? `No matching hidden ${pluralLabel}`
                      : `No hidden ${pluralLabel}`}
                  {/snippet}

                  {#snippet item({ item })}
                    {@const itemData = allItemsMap.get(item.id)}
                    {@const displayName =
                      itemData?.displayName ??
                      `Unknown ${type === "measure" ? "measure" : "dimension"}`}
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
                              aria-label={`Show ${displayName}`}
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
                          aria-label={`Show ${displayName}`}
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
          Clear the tag filter to reorder {pluralLabel}.
        </div>
      {:else if searchActive && !filterActive}
        <div
          class="px-3 py-1.5 text-xs text-fg-secondary border-t border-border"
        >
          Clear search to reorder {pluralLabel}.
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>
