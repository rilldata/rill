<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getComparisonOptionsForCanvas } from "@rilldata/web-common/features/canvas/filters/util";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { getComparisonInterval } from "@rilldata/web-common/lib/time/comparisons";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import type { DateTime, Interval } from "luxon";
  import {
    COMPARISON_RANGE_INHERIT,
    COMPARISON_RANGE_NONE,
    type ResolvedComparisonRange,
  } from "../../components/comparison-range";

  export let resolved: ResolvedComparisonRange;
  // What the canvas currently compares against, shown next to the inherit option.
  export let inheritedLabel: string;
  // The component's effective time range, which comparison options are computed against.
  export let interval: Interval<true> | undefined;
  export let selectedRangeAlias: string | undefined;
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let activeTimeZone: string;
  export let minTimeGrain: V1TimeGrain | undefined;
  export let minDate: DateTime<true> | undefined;
  export let maxDate: DateTime<true> | undefined;
  export let onSelect: (value: string) => void;

  let open = false;
  let showCustomPicker = false;

  $: selectedTimeRange = interval
    ? {
        name: selectedRangeAlias,
        start: interval.start.toJSDate(),
        end: interval.end.toJSDate(),
        interval: activeTimeGrain,
      }
    : undefined;

  $: options = getComparisonOptionsForCanvas(
    selectedTimeRange,
    false,
    activeTimeZone,
  );

  $: localRange = resolved.mode === "local" ? resolved.range : undefined;
  $: isCustom = localRange !== undefined && !(localRange in TIME_COMPARISON);
  $: customInterval = isCustom
    ? getComparisonInterval(interval, localRange, activeTimeZone)
    : undefined;

  $: label =
    resolved.mode === "inherit"
      ? m.canvas_comparison_inherit()
      : resolved.mode === "none"
        ? m.canvas_comparison_off()
        : isCustom
          ? m.time_custom_range()
          : (TIME_COMPARISON[localRange as TimeComparisonOption]?.label ??
            localRange);

  function select(value: string) {
    onSelect(value);
    open = false;
  }

  function applyCustomRange(range: Interval<true>) {
    select(`${range.start.toISO()},${range.end.toISO()}`);
  }
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={() => {
    showCustomPicker = isCustom;
  }}
>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <button
        {...props}
        type="button"
        aria-label={m.canvas_comparison_select_aria()}
        class="flex items-center gap-x-1.5 h-7 px-2 w-fit max-w-full rounded-sm border border-gray-300 bg-surface-card text-xs"
      >
        <b class="line-clamp-1">{label}</b>
        {#if resolved.mode === "inherit"}
          <span class="text-fg-secondary line-clamp-1">
            · {inheritedLabel}
          </span>
        {:else if customInterval}
          <RangeDisplay interval={customInterval} timeGrain={activeTimeGrain} />
        {/if}
        <span class="flex-none transition-transform" class:-rotate-180={open}>
          <CaretDownIcon />
        </span>
      </button>
    {/snippet}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="p-0 overflow-hidden">
    <div class="flex">
      <div class="flex flex-col border-r w-48 p-1">
        <DropdownMenu.Item onclick={() => select(COMPARISON_RANGE_INHERIT)}>
          <span class:font-bold={resolved.mode === "inherit"}>
            {m.canvas_comparison_inherit()}
          </span>
        </DropdownMenu.Item>
        <DropdownMenu.Item onclick={() => select(COMPARISON_RANGE_NONE)}>
          <span class:font-bold={resolved.mode === "none"}>
            {m.canvas_comparison_off()}
          </span>
        </DropdownMenu.Item>

        {#if options.length}
          <DropdownMenu.Separator />
        {/if}
        {#each options as option (option.name)}
          <DropdownMenu.Item onclick={() => select(option.name)}>
            <span class:font-bold={localRange === option.name}>
              {TIME_COMPARISON[option.name]?.label ?? option.name}
            </span>
          </DropdownMenu.Item>
        {/each}

        {#if interval}
          <DropdownMenu.Separator />
          <DropdownMenu.Item
            data-range="custom"
            closeOnSelect={false}
            onclick={() => {
              showCustomPicker = !showCustomPicker;
            }}
          >
            <span class:font-bold={isCustom}>{m.time_custom()}</span>
          </DropdownMenu.Item>
        {/if}
      </div>
      {#if showCustomPicker && interval}
        <div class="bg-surface-background flex flex-col w-60 p-3">
          <CalendarPlusDateInput
            minTimeGrain={V1TimeGrainToDateTimeUnit[
              minTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE
            ]}
            {maxDate}
            {minDate}
            interval={customInterval}
            zone={activeTimeZone}
            onApply={applyCustomRange}
            closeMenu={() => (open = false)}
          />
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
