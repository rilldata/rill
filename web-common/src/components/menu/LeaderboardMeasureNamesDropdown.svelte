<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { fly } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { writable } from "svelte/store";

  export let disabled = false;
  export let searchText = "";
  export let visibleMeasures: MetricsViewSpecMeasure[];
  export let leaderboardSortByMeasureName: string;
  export let selectedMeasureNames: string[];
  export let setLeaderboardMeasureNames: (names: string[]) => void;
  export let setLeaderboardSortByMeasureName: (name: string) => void;

  let active = false;
  let multiSelect = selectedMeasureNames?.length > 1;

  const lastSelectedMeasures = writable<string[]>([]);

  $: filteredMeasures = visibleMeasures
    .filter((item) =>
      ((item.displayName || item.name) ?? "")
        .toLowerCase()
        .includes(searchText.toLowerCase()),
    )
    .sort((a, b) => {
      const aIndex = visibleMeasures.findIndex((m) => m.name === a.name);
      const bIndex = visibleMeasures.findIndex((m) => m.name === b.name);
      return aIndex - bIndex;
    });

  $: showingMeasuresText =
    selectedMeasureNames.length > 1
      ? `${selectedMeasureNames.length} measures`
      : getMeasureDisplayText(leaderboardSortByMeasureName);

  function onToggleOff() {
    // Store the current selection before toggling off
    lastSelectedMeasures.set(selectedMeasureNames);

    // When toggling off multi-select, keep only the first visible measure
    if (selectedMeasureNames?.length > 0 && visibleMeasures[0]?.name) {
      const firstMeasure = visibleMeasures[0].name;
      setLeaderboardMeasureNames([firstMeasure]);
      setLeaderboardSortByMeasureName(firstMeasure);
    }
  }

  function toggleSingleSelect(name: string) {
    setLeaderboardSortByMeasureName(name);
    setLeaderboardMeasureNames([name]);
    active = false;
  }

  function toggleMultiSelect() {
    const newMultiSelect = !multiSelect;
    if (!newMultiSelect) {
      onToggleOff();
    } else {
      // When toggling back to multi-select, restore the last selection
      const lastMeasures = $lastSelectedMeasures;
      if (lastMeasures && lastMeasures.length > 0) {
        setLeaderboardMeasureNames(lastMeasures);
        setLeaderboardSortByMeasureName(lastMeasures[0]);
      }
    }
    multiSelect = newMultiSelect;
  }

  function toggleMeasure(name: string) {
    if (!name) return;
    const currentSelection = selectedMeasureNames || [];

    // Single select mode
    if (!multiSelect) {
      toggleSingleSelect(name);
      return;
    }

    // Multi-select mode
    const newSelection = currentSelection.includes(name)
      ? currentSelection.filter((n) => n !== name)
      : [...currentSelection, name];

    // Ensure we always have at least one measure selected
    if (newSelection.length === 0) {
      return;
    }

    setLeaderboardMeasureNames(newSelection);

    // If the toggled-off measure was the current sort measure
    // set the sort to the first remaining measure
    if (name === leaderboardSortByMeasureName && newSelection.length > 0) {
      setLeaderboardSortByMeasureName(newSelection[0]);
    }
  }

  function getMeasureDisplayText(measureName: string) {
    const measure = visibleMeasures.find((m) => m.name === measureName);
    return measure?.displayName || measure?.name || measureName;
  }
</script>

<DropdownMenu.Root
  closeOnItemClick={false}
  typeahead={false}
  bind:open={active}
  onOpenChange={(open) => {
    if (!open) searchText = "";
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={active}
    >
      <Button
        builders={[builder]}
        type="text"
        theme
        dataAttributes={{
          "data-testid": "leaderboard-measure-names-dropdown",
          "data-leaderboard-measures-count":
            selectedMeasureNames.length.toString(),
        }}
      >
        <div
          class="flex items-center gap-x-0.5 px-1 text-surface-foreground dark:text-muted-foreground hover:text-inherit"
        >
          Showing <strong> {showingMeasuresText}</strong>
          <span
            class="transition-transform"
            class:hidden={disabled}
            class:-rotate-180={active}
          >
            <CaretDownIcon />
          </span>
        </div>
      </Button>

      <DropdownMenu.Content
        align="start"
        class="flex flex-col w-72 p-0 overflow-hidden"
        strategy="absolute"
        fitViewport={true}
      >
        <div class="px-3 pt-3 pb-1">
          <Search
            bind:value={searchText}
            label="Search measures"
            showBorderOnFocus={false}
          />
        </div>

        <div class="px-1 pb-1 max-h-80 overflow-y-auto">
          {#if filteredMeasures.length}
            {#each filteredMeasures as measure (measure.name)}
              <DropdownMenu.CheckboxItem
                class="text-[12px]"
                checked={Boolean(
                  measure.name &&
                    (multiSelect
                      ? selectedMeasureNames.includes(measure.name)
                      : leaderboardSortByMeasureName === measure.name),
                )}
                onCheckedChange={() => {
                  if (measure.name) toggleMeasure(measure.name);
                }}
              >
                <div class="truncate flex-1 text-left">
                  {measure.displayName || measure.name}
                </div>
              </DropdownMenu.CheckboxItem>
            {/each}
          {:else}
            <div class="ui-copy-disabled p-2 w-full">
              No matching leaderboard measures shown
            </div>
          {/if}
        </div>

        {#if visibleMeasures.length > 1}
          <footer class="bg-popover-footer">
            <div class="flex items-center space-x-2">
              <Switch
                theme
                checked={multiSelect}
                id="multi-measure-select"
                small
                on:click={toggleMultiSelect}
                data-testid="multi-measure-select-switch"
              />
              <InputLabel
                small
                capitalize={false}
                label="Multi-select"
                id="multi-measure-select"
              />
            </div>
          </footer>
        {/if}
      </DropdownMenu.Content>

      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          Choose measures to display
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>
</DropdownMenu.Root>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t;
    @apply flex flex-row flex-none items-center justify-start;
    @apply gap-x-2 p-2 px-3.5;
  }
</style>
