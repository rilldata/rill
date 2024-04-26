<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";

  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import { CustomEventHandler } from "bits-ui";
  import {
    RILL_TO_LABEL,
    ALL_TIME_RANGE_ALIAS,
    getRangeLabel,
  } from "../../new-time-controls";
  import TimeRangeMenu from "./TimeRangeMenu.svelte";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;

  export let showDefaultItem: boolean;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;

  let open = false;
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger aria-label="Select a time range" asChild let:builder>
    <button {...builder} use:builder.action>
      {getRangeLabel(selected)}
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
    @apply font-bold px-3;
  }
</style>
