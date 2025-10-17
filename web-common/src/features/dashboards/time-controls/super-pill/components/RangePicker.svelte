<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime, Interval } from "luxon";
  import type {
    ISODurationString,
    NamedRange,
    RangeBuckets,
  } from "../../new-time-controls";
  import { getRangeLabel } from "../../new-time-controls";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";
  import RangeDisplay from "./RangeDisplay.svelte";
  import TimeRangeMenu from "./TimeRangeMenu.svelte";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let interval: Interval<true>;
  export let zone: string;
  export let showDefaultItem: boolean;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let showFullRange: boolean;
  export let allowCustomTimeRange = true;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let applyCustomRange: (range: Interval<true>) => void;

  let open = false;
  let showSelector = false;
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={() => {
    showSelector = selected === "CUSTOM";
  }}
  closeOnItemClick={false}
  typeahead={!showSelector}
>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex gap-x-1"
      aria-label="Select time range"
      type="button"
    >
      <b class="mr-1 line-clamp-1 flex-none">{getRangeLabel(selected)}</b>

      {#if interval.isValid && showFullRange}
        <RangeDisplay {interval} />
      {/if}
      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" {side} class="p-0 overflow-hidden">
    <div class="flex">
      <div class="flex flex-col w-48 p-1">
        <TimeRangeMenu
          {ranges}
          {selected}
          {showDefaultItem}
          {defaultTimeRange}
          {allowCustomTimeRange}
          onSelectRange={(selected) => {
            onSelectRange(selected);

            open = false;
          }}
          onSelectCustomOption={() => (showSelector = !showSelector)}
        />
      </div>
      {#if showSelector}
        <div class="bg-slate-50 border-l flex flex-col w-64 p-3">
          <CalendarPlusDateInput
            {interval}
            {zone}
            {maxDate}
            {minDate}
            minTimeGrain={V1TimeGrainToDateTimeUnit[
              smallestTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE
            ]}
            onApply={(i) => {
              applyCustomRange(i);
            }}
            closeMenu={() => (open = false)}
          />
        </div>
      {/if}
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>
