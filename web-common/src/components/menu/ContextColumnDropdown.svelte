<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { fly } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
  import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
  import PercentOfTotal from "@rilldata/web-common/features/dashboards/dimension-table/PercentOfTotal.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  export let isValidPercentOfTotal: boolean;
  export let tooltipText: string;
  export let measures: MetricsViewSpecMeasureV2[];
  export let selectedFilters: LeaderboardContextColumn[] = [];
  export let selectedMeasureNames: string[] = [];
  export let onToggle: (column: LeaderboardContextColumn[]) => void;
  export let onSelectAll: () => void;

  const { exploreName } = getStateManagers();

  let active = false;

  function removeTimeComparisonColumns(filters: LeaderboardContextColumn[]) {
    return filters.filter(
      (f) =>
        f !== LeaderboardContextColumn.DELTA_ABSOLUTE &&
        f !== LeaderboardContextColumn.DELTA_PERCENT &&
        f !== LeaderboardContextColumn.PERCENT,
    );
  }

  // Side effect to clean up filters when time comparison is disabled
  $: if (!$metricsExplorerStore.entities[$exploreName]?.showTimeComparison) {
    const cleanedFilters = removeTimeComparisonColumns(selectedFilters);
    if (cleanedFilters.length !== selectedFilters.length) {
      onToggle(cleanedFilters);
    }
  }

  $: options = [
    ...(isValidPercentOfTotal
      ? [
          {
            value: LeaderboardContextColumn.PERCENT,
            label: "Percent of total",
            description: "Summable metrics only",
            icon: PercentOfTotal,
          },
        ]
      : []),
    {
      value: LeaderboardContextColumn.DELTA_ABSOLUTE,
      label: "Change",
      icon: DeltaChange,
    },
    {
      value: LeaderboardContextColumn.DELTA_PERCENT,
      label: "Percent change",
      icon: DeltaChangePercentage,
    },
  ];

  function getLabelFromValue(value: LeaderboardContextColumn) {
    return options.find((option) => option.value === value)?.label;
  }

  function toggleContextColumn(name: string) {
    if (!name) return;
    const column = name as LeaderboardContextColumn;
    const isAdding = !selectedFilters.includes(column);
    const newFilters = isAdding
      ? [...selectedFilters, column]
      : selectedFilters.filter((f) => f !== column);
    onToggle(newFilters);

    // If adding a delta column and comparison time range is not enabled,
    // automatically enable it with a default comparison range
    if (
      isAdding &&
      (column === LeaderboardContextColumn.DELTA_ABSOLUTE ||
        column === LeaderboardContextColumn.DELTA_PERCENT ||
        column === LeaderboardContextColumn.PERCENT) &&
      !$metricsExplorerStore.entities[$exploreName]?.showTimeComparison
    ) {
      const defaultComparisonRange = {
        name: TimeComparisonOption.CONTIGUOUS,
        start: new Date(),
        end: new Date(),
      };
      const currentMeasures = measures.map((m) => ({ name: m.name }));
      metricsExplorerStore.setSelectedComparisonRange(
        $exploreName,
        defaultComparisonRange,
        { measures: currentMeasures },
      );
      metricsExplorerStore.displayTimeComparison($exploreName, true);
    }
  }

  $: allSelected = selectedMeasureNames.length === measures.length;

  $: withText =
    selectedFilters && selectedFilters.length > 1
      ? `${selectedFilters.length} context columns`
      : selectedFilters.length === 1
        ? getLabelFromValue(selectedFilters[0])
        : "no context columns";
</script>

<DropdownMenu.Root
  closeOnItemClick={false}
  typeahead={false}
  bind:open={active}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={active}
    >
      <Button builders={[builder]} type="text" on:click>
        <div
          class="flex items-center gap-x-1 px-1 text-gray-700 hover:text-inherit font-normal"
        >
          with <strong>{withText}</strong>
          <span class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon />
          </span>
        </div>
      </Button>

      <DropdownMenu.Content
        align="start"
        class="flex flex-col max-h-96 w-[204px] overflow-hidden p-0"
      >
        <div class="px-1 pb-1 pt-1">
          {#each options as option}
            <DropdownMenu.CheckboxItem
              checked={selectedFilters.includes(option.value)}
              onCheckedChange={() => toggleContextColumn(option.value)}
            >
              <div class="flex items-center">
                {#if option.value === LeaderboardContextColumn.DELTA_ABSOLUTE}
                  <div class="flex items-center justify-start w-[26px]">
                    <svelte:component this={option.icon} />
                  </div>
                  <span>{option.label}</span>
                {:else if option.value === LeaderboardContextColumn.DELTA_PERCENT}
                  <div class="flex items-center justify-start w-[26px]">
                    <svelte:component this={option.icon} />
                  </div>
                  <span>{option.label}</span>
                {:else if option.value === LeaderboardContextColumn.PERCENT}
                  <div class="flex flex-col">
                    <div class="flex flex-row gap-x-1">
                      <svelte:component this={option.icon} />
                      <span>{option.label}</span>
                    </div>
                    <span class="ui-copy-muted text-[11px]">
                      {option.description}
                    </span>
                  </div>
                {/if}
              </div>
            </DropdownMenu.CheckboxItem>
          {/each}
        </div>

        <footer>
          <div class="w-full">
            <p class="text-xs">Show for all measures</p>
          </div>
          <Switch small bind:checked={allSelected} on:click={onSelectAll} />
        </footer>
      </DropdownMenu.Content>

      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          {tooltipText}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>
</DropdownMenu.Root>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-white px-3.5 py-2;
    @apply flex flex-row flex-none items-center gap-x-2;
  }
</style>
