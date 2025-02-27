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

  export let isValidPercentOfTotal = false;
  export let isTimeComparisonActive = false;
  export let tooltipText: string;
  export let selectedFilters: LeaderboardContextColumn[] = [];
  export let onContextColumnChange: (
    column: LeaderboardContextColumn[],
  ) => void;

  let active = false;

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
    ...(isTimeComparisonActive
      ? [
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
        ]
      : []),
  ];

  function getLabelFromValue(value: LeaderboardContextColumn) {
    return options.find((option) => option.value === value)?.label;
  }

  function toggleContextColumn(name: string) {
    if (!name) return;
    const column = name as LeaderboardContextColumn;
    const newFilters = selectedFilters.includes(column)
      ? selectedFilters.filter((f) => f !== column)
      : [...selectedFilters, column];
    onContextColumnChange(newFilters);
  }

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
      </DropdownMenu.Content>

      <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
        <TooltipContent maxWidth="400px">
          {tooltipText}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>
</DropdownMenu.Root>
