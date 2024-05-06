<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import { getRangeLabel } from "../../new-time-controls";
  import TimeRangeMenu from "./TimeRangeMenu.svelte";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { Interval } from "luxon";
  import RangeDisplay from "./RangeDisplay.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let interval: Interval;
  export let grain: V1TimeGrain | undefined;
  export let showDefaultItem: boolean;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;

  let open = false;
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger aria-label="Select a time range" asChild let:builder>
    <button {...builder} use:builder.action class="flex gap-x-1">
      <b class="mr-1 line-clamp-1 flex-none">{getRangeLabel(selected)}</b>
      {#if interval.isValid}
        <RangeDisplay {interval} {grain} />
      {/if}
      <CaretDownIcon />
    </button>
  </DropdownMenu.Trigger>
  <TimeRangeMenu
    {ranges}
    {selected}
    {showDefaultItem}
    {defaultTimeRange}
    {onSelectRange}
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
  }
</style>
