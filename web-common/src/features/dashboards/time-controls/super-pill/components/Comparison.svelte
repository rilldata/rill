<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { DateTime, Interval } from "luxon";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";
  import RangeDisplay from "./RangeDisplay.svelte";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import WarningIcon from "@rilldata/web-common/components/icons/WarningIcon.svelte";
  import { Tooltip } from "bits-ui";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getComparisonRange } from "@rilldata/web-common/lib/time/comparisons";

  export let comparisonOptions: TimeComparisonOption[];
  export let showComparison: boolean | undefined;
  export let currentInterval: Interval<true> | undefined;
  export let comparisonInterval: Interval<true> | undefined;
  export let comparisonRange: string | undefined;
  export let grain: V1TimeGrain | undefined;
  export let zone: string;
  export let disabled: boolean;
  export let showFullRange: boolean;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let allowCustomTimeRange: boolean = true;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let onSelectComparisonString: ((name: string) => void) | undefined =
    undefined;
  export let onSelectComparisonRange:
    | ((name: string, start: Date, end: Date) => void)
    | undefined = undefined;

  let open = false;
  let showSelector = false;

  $: firstVisibleMonth =
    comparisonInterval?.start ?? currentInterval?.end ?? DateTime.now();

  $: firstOption = comparisonOptions[0];
  $: label =
    TIME_COMPARISON[comparisonRange ?? firstOption]?.label ?? "Custom range";

  $: selectedLabel = comparisonRange ?? firstOption ?? "Custom range";

  $: intervalsOverlap =
    currentInterval && comparisonInterval?.overlaps(currentInterval);
  $: comparisonIntervalIsOutsideOfRange =
    showComparison &&
    comparisonRange &&
    comparisonInterval &&
    minDate &&
    maxDate &&
    (comparisonInterval.end <= minDate || comparisonInterval.start >= maxDate);

  function applyRange(range: Interval<true>) {
    const string = `${range.start.toISODate()},${range.end.toISODate()}`;
    if (onSelectComparisonString) {
      onSelectComparisonString?.(string);
    } else if (onSelectComparisonRange) {
      onSelectComparisonRange?.(
        TimeComparisonOption.CUSTOM,
        range.start.toJSDate(),
        range.end.toJSDate(),
      );
    }
  }

  function onCompareRangeSelect(comparisonOption: TimeComparisonOption) {
    if (
      currentInterval?.isValid &&
      currentInterval.start &&
      currentInterval.end
    ) {
      const comparisonTimeRange = getComparisonRange(
        currentInterval.start.toJSDate(),
        currentInterval.end.toJSDate(),
        comparisonOption,
      );

      onSelectComparisonRange?.(
        comparisonOption,
        comparisonTimeRange.start,
        comparisonTimeRange.end,
      );
    }
  }
</script>

<DropdownMenu.Root
  bind:open
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open && comparisonInterval && comparisonInterval?.isValid) {
      firstVisibleMonth = comparisonInterval.start;
    }
    showSelector = !!(
      comparisonRange === TimeComparisonOption.CUSTOM && showComparison
    );
  }}
  typeahead={!showSelector}
>
  <DropdownMenu.Trigger asChild let:builder {disabled}>
    <button
      {disabled}
      aria-disabled={disabled}
      use:builder.action
      {...builder}
      aria-label="Select time comparison option"
      type="button"
    >
      <div class="gap-x-2 flex" class:opacity-50={!showComparison}>
        {#if !comparisonOptions.length && !showComparison}
          <p>no comparison period</p>
        {:else}
          <b class="line-clamp-1">{label}</b>
          {#if comparisonInterval?.isValid && showFullRange}
            <RangeDisplay interval={comparisonInterval} timeGrain={grain} />
          {/if}
        {/if}
        {#if intervalsOverlap || comparisonIntervalIsOutsideOfRange}
          <Tooltip.Root portal="#rill-portal">
            <Tooltip.Trigger>
              <WarningIcon className="text-yellow-500" />
            </Tooltip.Trigger>
            <Tooltip.Content class="z-50" sideOffset={8}>
              <TooltipContent>
                {#if comparisonIntervalIsOutsideOfRange}
                  No data for comparison period
                {:else if intervalsOverlap}
                  The selected comparison range overlaps with the primary range.
                {/if}
              </TooltipContent>
            </Tooltip.Content>
          </Tooltip.Root>
        {/if}
      </div>
      <span
        class="flex-none transition-transform"
        class:-rotate-180={open}
        class:opacity-50={!showComparison}
      >
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" {side} class="p-0 overflow-hidden">
    <div class="flex">
      <div class="flex flex-col border-r w-48 p-1">
        {#each comparisonOptions as option (option)}
          {@const preset = TIME_COMPARISON[option]}
          {@const selected = selectedLabel === option}
          <DropdownMenu.Item
            class="flex gap-x-2"
            on:click={() => {
              if (onSelectComparisonString) {
                onSelectComparisonString(option);
              }
              if (onSelectComparisonRange) {
                onCompareRangeSelect(option);
              }
              open = false;
            }}
          >
            <span class:font-bold={selected}>
              {preset?.label || option}
            </span>
          </DropdownMenu.Item>
          {#if option === TimeComparisonOption.CONTIGUOUS && comparisonOptions.length > 2}
            <DropdownMenu.Separator />
          {/if}
        {/each}
        {#if allowCustomTimeRange}
          {#if comparisonOptions.length}
            <DropdownMenu.Separator />
          {/if}

          <DropdownMenu.Item
            data-range="custom"
            on:click={() => {
              showSelector = !showSelector;
            }}
          >
            <span
              class:font-bold={comparisonRange ===
                TimeComparisonOption.CUSTOM && showComparison}
            >
              Custom
            </span>
          </DropdownMenu.Item>
        {/if}
      </div>
      {#if showSelector}
        <div class="bg-slate-50 flex flex-col w-64 px-2 py-1">
          {#if !comparisonInterval || comparisonInterval?.isValid}
            <CalendarPlusDateInput
              {maxDate}
              {minDate}
              {firstVisibleMonth}
              interval={comparisonInterval}
              {zone}
              {applyRange}
              closeMenu={() => (open = false)}
            />
          {/if}
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply gap-x-1;
  }

  .inactive {
    @apply opacity-50;
  }
</style>
