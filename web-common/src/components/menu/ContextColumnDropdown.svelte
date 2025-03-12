<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { fly } from "svelte/transition";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
  import DeltaChange from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChange.svelte";
  import DeltaChangePercentage from "@rilldata/web-common/features/dashboards/dimension-table/DeltaChangePercentage.svelte";
  import PercentOfTotal from "@rilldata/web-common/features/dashboards/dimension-table/PercentOfTotal.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  export let atLeastOneValidPercentOfTotal: boolean;
  export let isTimeComparisonActive: boolean;
  export let tooltipText: string;
  export let contextColumns: string[] | undefined;
  export let dimensionShowAllMeasures: boolean;
  export let onContextColumnChange: (
    columns: LeaderboardContextColumn[],
  ) => void;
  export let onShowForAllMeasures: () => void;

  let active = false;

  function shouldSuppress(option) {
    const isPercentOption = option.value === LeaderboardContextColumn.PERCENT;
    const isDisabled = !atLeastOneValidPercentOfTotal;
    return !(isPercentOption && isDisabled);
  }

  $: options = [
    {
      value: LeaderboardContextColumn.PERCENT,
      label: "Percent of total",
      description: "Summable metrics only",
      icon: PercentOfTotal,
    },
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

  function getLabelFromValue(value: string) {
    return (
      options.find((option) => option.value === value)?.label ||
      "unknown context column"
    );
  }

  function toggleContextColumn(name: string) {
    if (!name) return;
    const column = name as LeaderboardContextColumn;
    const newFilters = contextColumns?.includes(column)
      ? contextColumns?.filter((f) => f !== column)
      : [...(contextColumns || []), column];
    onContextColumnChange(newFilters as LeaderboardContextColumn[]);
  }

  // WORKAROUND for when comparison is off and have delta or percent context columns selected
  // Remove time comparison columns when time comparison is disabled
  $: {
    if (!isTimeComparisonActive) {
      const filteredColumns = contextColumns?.filter(
        (column) =>
          column !== LeaderboardContextColumn.DELTA_ABSOLUTE &&
          column !== LeaderboardContextColumn.DELTA_PERCENT,
      );
      if (filteredColumns?.length !== contextColumns?.length) {
        onContextColumnChange(filteredColumns as LeaderboardContextColumn[]);
      }
    }
  }

  $: withText = !contextColumns
    ? "no context columns"
    : contextColumns.length > 1
      ? `${contextColumns.length} context columns`
      : contextColumns.length === 1
        ? getLabelFromValue(contextColumns[0])
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
        {#if options.length > 0}
          <div class="px-1 pb-1 pt-1">
            {#each options as option}
              <Tooltip
                distance={8}
                suppress={shouldSuppress(option)}
                location="right"
                alignment="middle"
              >
                <DropdownMenu.CheckboxItem
                  checked={contextColumns?.includes(option.value)}
                  onCheckedChange={() => toggleContextColumn(option.value)}
                  disabled={!atLeastOneValidPercentOfTotal &&
                    option.value === LeaderboardContextColumn.PERCENT}
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
                <TooltipContent maxWidth="400px" slot="tooltip-content">
                  Only available for metrics marked as summable
                </TooltipContent>
              </Tooltip>
            {/each}
          </div>
        {/if}

        <footer class={cn(options.length > 0 && "border-t border-slate-300")}>
          <div class="w-full">
            <p class="text-xs">Show for all measures</p>
          </div>
          <Switch
            small
            bind:checked={dimensionShowAllMeasures}
            on:click={onShowForAllMeasures}
          />
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
    @apply bg-white px-3.5 py-2;
    @apply flex flex-row flex-none items-center gap-x-2;
  }
</style>
