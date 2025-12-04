<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import type { CustomEventHandler } from "bits-ui";
  import { RILL_TO_LABEL, ALL_TIME_RANGE_ALIAS } from "../../new-time-controls";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let showDefaultItem: boolean;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let onSelectCustomOption: () => void;
  export let allowCustomTimeRange = true;

  function handleClick(e: CustomEventHandler<MouseEvent, HTMLDivElement>) {
    const range = e.detail.currentTarget.dataset.range;
    if (!range) {
      throw new Error("No range provided");
    }

    onSelectRange(range);
  }
</script>

{#if showDefaultItem && defaultTimeRange}
  <DropdownMenu.Item data-range={defaultTimeRange} on:click={handleClick}>
    <div class:font-bold={selected === defaultTimeRange}>
      Last {humaniseISODuration(defaultTimeRange)}
    </div>
  </DropdownMenu.Item>

  <DropdownMenu.Separator />
{/if}

{#each ranges.latest as rillTime, i (i)}
  <DropdownMenu.Item
    data-range={rillTime.interval.toString()}
    on:click={handleClick}
  >
    <span class:font-bold={selected === rillTime.interval.toString()}>
      {rillTime.getLabel()}
    </span>
  </DropdownMenu.Item>
{/each}

{#if ranges.latest.length}
  <DropdownMenu.Separator />
{/if}

{#each ranges.periodToDate as rillTime, i (i)}
  <DropdownMenu.Item
    data-range={rillTime.interval.toString()}
    on:click={handleClick}
  >
    <span class:font-bold={selected === rillTime.interval.toString()}>
      {rillTime.getLabel()}
    </span>
  </DropdownMenu.Item>
{/each}

{#if ranges.periodToDate.length}
  <DropdownMenu.Separator />
{/if}

{#each ranges.previous as rillTime, i (i)}
  <DropdownMenu.Item
    data-range={rillTime.interval.toString()}
    on:click={handleClick}
  >
    <span class:font-bold={selected === rillTime.interval.toString()}>
      {rillTime.getLabel()}
    </span>
  </DropdownMenu.Item>
{/each}

{#if ranges.allTime}
  <DropdownMenu.Separator />
  <DropdownMenu.Item on:click={handleClick} data-range={ALL_TIME_RANGE_ALIAS}>
    <span class:font-bold={selected === ALL_TIME_RANGE_ALIAS}>
      {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
    </span>
  </DropdownMenu.Item>
{/if}

{#if allowCustomTimeRange}
  <DropdownMenu.Separator />
  <DropdownMenu.Item on:click={onSelectCustomOption} data-range="custom">
    <span class:font-bold={selected === "CUSTOM"}> Custom </span>
  </DropdownMenu.Item>
{/if}
