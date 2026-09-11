<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types.ts";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config.ts";
  import {
    MinSupportedGrain,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains.ts";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import type { Interval } from "luxon";
  import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

  let {
    timeFilterManager,
    metricsViewsProvider,
    config,
  }: {
    timeFilterManager: TimeFilterManager;
    metricsViewsProvider: MetricsViewsProvider;
    config: TimeFiltersConfig;
  } = $props();
  let {
    showFullRange = true,
    allowCustomTimeRange = true,
    side = "bottom",
  } = $derived(config);

  let {
    timeRange,
    timeGrain,
    timeZone,
    minDate,
    maxDate,
    comparisonTimeRange,
    showComparison,
    comparisonTimeRangeOptions,
    comparisonInterval,
  } = $derived(timeFilterManager);

  let { smallestTimeGrain } = $derived(metricsViewsProvider);

  let open = $state(false);
  let showSelector = $state(false);

  let disabled = $derived(timeRange === TimeRangePreset.ALL_TIME || undefined);

  let firstOption = $derived(comparisonTimeRangeOptions[0]);
  let label = $derived(
    TIME_COMPARISON[comparisonTimeRange ?? firstOption?.name]?.label ??
      m.time_custom_range(),
  );
  let selectedLabel = $derived(
    comparisonTimeRange ?? firstOption?.name ?? m.time_custom_range(),
  );

  function applyCustomRange(range: Interval<true>) {
    onSelectComparisonRange(`${range.start.toISO()} to ${range.end.toISO()}`);
  }

  function onSelectComparisonRange(range: string) {
    open = false;
    timeFilterManager.onSelectComparisonRange(range);
  }
</script>

<div
  class="wrapper"
  title={disabled && "Comparison not available when viewing all time range"}
>
  <button
    {disabled}
    class="flex gap-x-1.5 cursor-pointer"
    onclick={() => timeFilterManager.onToggleShowComparison()}
    type="button"
    aria-label={m.dashboard_toggle_time_comparison_aria()}
  >
    <div class="pointer-events-none flex items-center gap-x-1.5">
      <Switch
        checked={showComparison}
        id="comparing"
        small
        disabled={disabled ?? false}
      />

      <Label class="font-normal text-xs cursor-pointer" for="comparing">
        <span class:opacity-50={disabled}>{m.time_comparing()}</span>
      </Label>
    </div>
  </button>
  {#if timeGrain && comparisonInterval}
    <DropdownMenu.Root
      bind:open
      onOpenChange={() => {
        showSelector = !!(
          comparisonTimeRange === TimeComparisonOption.CUSTOM && showComparison
        );
      }}
    >
      <DropdownMenu.Trigger {disabled}>
        {#snippet child({ props })}
          <button
            {...props}
            {disabled}
            aria-disabled={disabled}
            aria-label={m.dashboard_select_time_comparison_aria()}
            type="button"
            class="rounded-r-full"
          >
            <div class="gap-x-2 flex" class:opacity-50={!showComparison}>
              {#if !comparisonTimeRangeOptions.length && !showComparison}
                <p>{m.time_no_comparison_period()}</p>
              {:else}
                <b class="line-clamp-1">{label}</b>
                {#if comparisonInterval?.isValid && showFullRange}
                  <RangeDisplay interval={comparisonInterval} {timeGrain} />
                {/if}
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
        {/snippet}
      </DropdownMenu.Trigger>

      <DropdownMenu.Content align="start" {side} class="p-0 overflow-hidden">
        <div class="flex">
          <div class="flex flex-col border-r w-48 p-1">
            {#each comparisonTimeRangeOptions as option (option.name)}
              {@const preset = TIME_COMPARISON[option.name]}
              {@const selected = selectedLabel === option.name}
              <DropdownMenu.Item
                class="flex gap-x-2"
                onclick={() => onSelectComparisonRange(option.name)}
              >
                <span class:font-bold={selected}>
                  {preset?.label || option.name}
                </span>
              </DropdownMenu.Item>
              {#if option.name === TimeComparisonOption.CONTIGUOUS && comparisonTimeRangeOptions.length > 2}
                <DropdownMenu.Separator />
              {/if}
            {/each}
            {#if allowCustomTimeRange}
              {#if comparisonTimeRangeOptions.length}
                <DropdownMenu.Separator />
              {/if}

              <DropdownMenu.Item
                data-range="custom"
                closeOnSelect={false}
                onclick={() => {
                  showSelector = !showSelector;
                }}
              >
                <span
                  class:font-bold={comparisonTimeRange ===
                    TimeComparisonOption.CUSTOM && showComparison}
                >
                  {m.time_custom()}
                </span>
              </DropdownMenu.Item>
            {/if}
          </div>
          {#if showSelector}
            <div class="bg-surface-background flex flex-col w-60 p-3">
              {#if !comparisonInterval || comparisonInterval?.isValid}
                <CalendarPlusDateInput
                  minTimeGrain={V1TimeGrainToDateTimeUnit[
                    smallestTimeGrain ?? MinSupportedGrain
                  ]}
                  {maxDate}
                  {minDate}
                  interval={comparisonInterval}
                  zone={timeZone}
                  onApply={applyCustomRange}
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
    </style>
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-[26px] rounded-full;
    @apply overflow-hidden select-none;
  }
</style>
