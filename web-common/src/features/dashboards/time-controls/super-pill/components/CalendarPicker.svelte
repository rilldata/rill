<script context="module" lang="ts">
  import { writable } from "svelte/store";
  import { Interval } from "luxon";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CalendarIcon from "@rilldata/web-common/components/icons/Calendar.svelte";
  import { DateTime } from "luxon";
  import CalendarPlusDateInput from "./CalendarPlusDateInput.svelte";

  export const open = writable(false);
</script>

<script lang="ts">
  export let interval: Interval<true>;
  export let zone: string;
  export let applyRange: (range: Interval<true>) => void;

  let firstVisibleMonth: DateTime<true> = interval.start;
</script>

<DropdownMenu.Root
  bind:open={$open}
  onOpenChange={(open) => {
    if (open) {
      firstVisibleMonth = interval.start;
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      use:builder.action
      {...builder}
      aria-label="Select a custom time range"
    >
      <CalendarIcon size="16px" />
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-72">
    <CalendarPlusDateInput
      {firstVisibleMonth}
      {interval}
      {zone}
      {applyRange}
      closeMenu={() => open.set(false)}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
    /* this resolves an issue in safari */
    @apply transform;
  }
</style>
