<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { fly } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
  import PieChart from "@rilldata/web-common/components/icons/PieChart.svelte";

  export let isValidPercentOfTotal = false;
  export let isTimeComparisonActive = false;
  export let tooltipText: string;
  export let selected: LeaderboardContextColumn;
  export let onContextColumnChange: (column: LeaderboardContextColumn) => void;

  let active = false;

  // TODO: support multi select
  // TODO: look into the relationship between context column and all measures
  // TODO: will we have context columns synced with the url state?

  $: options = [
    ...(isTimeComparisonActive
      ? [
          {
            value: LeaderboardContextColumn.DELTA_ABSOLUTE,
            label: "Absolute change",
          },
          {
            value: LeaderboardContextColumn.DELTA_PERCENT,
            label: "Percent change",
          },
        ]
      : []),
    ...(isValidPercentOfTotal
      ? [
          {
            value: LeaderboardContextColumn.PERCENT,
            label: "Percent of total",
            description: "Summable metrics only",
          },
        ]
      : []),
  ];

  function getLabelFromValue(value: LeaderboardContextColumn) {
    return options.find((option) => option.value === value)?.label;
  }

  function toggleContextColumn(name: string) {
    if (!name) return;
    onContextColumnChange(name as LeaderboardContextColumn);
  }

  $: withText =
    selected !== LeaderboardContextColumn.HIDDEN
      ? getLabelFromValue(selected)
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
          class="flex items-center gap-x-0.5 px-1 text-gray-700 hover:text-inherit"
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
              checked={selected === option.value}
              onCheckedChange={() => toggleContextColumn(option.value)}
            >
              <div class="flex items-center gap-x-1">
                <div class="flex items-center justify-start">
                  {#if option.value === LeaderboardContextColumn.DELTA_ABSOLUTE}
                    <Delta />
                    <div class="w-4" />
                  {:else if option.value === LeaderboardContextColumn.DELTA_PERCENT}
                    <Delta />
                    <div class="w-4">%</div>
                  {:else if option.value === LeaderboardContextColumn.PERCENT}
                    <PieChart />
                    <div class="w-4">%</div>
                  {/if}
                </div>
                {option.label}
                {#if option.description}
                  <span class="ui-copy-muted text-[11px] ml-1">
                    ({option.description})
                  </span>
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

<style lang="postcss">
</style>
