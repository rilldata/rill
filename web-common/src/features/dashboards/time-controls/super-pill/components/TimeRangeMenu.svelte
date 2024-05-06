<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import { open as calendarOpen } from "./CalendarPicker.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import { CustomEventHandler } from "bits-ui";
  import { RILL_TO_LABEL, ALL_TIME_RANGE_ALIAS } from "../../new-time-controls";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;

  export let showDefaultItem: boolean;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;

  function handleClick(e: CustomEventHandler<MouseEvent, HTMLDivElement>) {
    const range = e.detail.currentTarget.dataset.range;
    if (!range) {
      throw new Error("No range provided");
    }

    onSelectRange(range);
  }

  function handleCustom() {
    calendarOpen.set(true);
  }
</script>

<DropdownMenu.Content align="start" class="w-80">
  <DropdownMenu.Item on:click={handleClick} data-range={ALL_TIME_RANGE_ALIAS}>
    <span class:font-bold={selected === ALL_TIME_RANGE_ALIAS}>
      {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
    </span>
  </DropdownMenu.Item>
  {#if showDefaultItem && defaultTimeRange}
    <DropdownMenu.Item data-range={defaultTimeRange} on:click={handleClick}>
      <div class:font-bold={selected === defaultTimeRange}>
        Last {humaniseISODuration(defaultTimeRange)}
      </div>
    </DropdownMenu.Item>
  {/if}

  {#if ranges.latest.length}
    <DropdownMenu.Separator />
  {/if}

  {#each ranges.latest as { range, label } (range)}
    <DropdownMenu.Item on:click={handleClick} data-range={range}>
      <span class:font-bold={selected === range}>
        {label}
      </span>
    </DropdownMenu.Item>
  {/each}

  {#if ranges.periodToDate.length}
    <DropdownMenu.Separator />
  {/if}

  {#each ranges.periodToDate as { range, label } (range)}
    <DropdownMenu.Item on:click={handleClick} data-range={range}>
      <span class:font-bold={selected === range}>
        {label}
      </span>
    </DropdownMenu.Item>
  {/each}

  {#if ranges.previous.length}
    <DropdownMenu.Separator />
  {/if}

  {#each ranges.previous as { range, label } (range)}
    <DropdownMenu.Item on:click={handleClick} data-range={range}>
      <span class:font-bold={selected === range}>
        {label}
      </span>
    </DropdownMenu.Item>
  {/each}

  <DropdownMenu.Separator />

  <DropdownMenu.Item on:click={handleCustom}>
    <span class:font-bold={selected === "Custom"}> Custom </span>
  </DropdownMenu.Item>
</DropdownMenu.Content>
