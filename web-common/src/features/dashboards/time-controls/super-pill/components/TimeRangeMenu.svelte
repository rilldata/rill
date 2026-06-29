<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
  } from "../../new-time-controls";
  import { RILL_TO_LABEL, ALL_TIME_RANGE_ALIAS } from "../../new-time-controls";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let showDefaultItem: boolean;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let onSelectCustomOption: () => void;
  export let allowCustomTimeRange = true;

  function handleClick(e: MouseEvent) {
    const range = (e.currentTarget as HTMLElement)?.dataset?.range;
    if (!range) {
      throw new Error("No range provided");
    }

    onSelectRange(range);
  }
</script>

{#if showDefaultItem && defaultTimeRange}
  <DropdownMenu.Item data-range={defaultTimeRange} onclick={handleClick}>
    <div class:font-bold={selected === defaultTimeRange}>
      {m.time_last_duration({
        duration: humaniseISODuration(defaultTimeRange),
      })}
    </div>
  </DropdownMenu.Item>

  <DropdownMenu.Separator />
{/if}

{#each ranges.latest as rillTime, i (i)}
  <DropdownMenu.Item
    data-range={rillTime.interval.toString()}
    onclick={handleClick}
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
    onclick={handleClick}
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
    onclick={handleClick}
  >
    <span class:font-bold={selected === rillTime.interval.toString()}>
      {rillTime.getLabel()}
    </span>
  </DropdownMenu.Item>
{/each}

{#if ranges.allTime}
  <DropdownMenu.Separator />
  <DropdownMenu.Item onclick={handleClick} data-range={ALL_TIME_RANGE_ALIAS}>
    <span class:font-bold={selected === ALL_TIME_RANGE_ALIAS}>
      {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
    </span>
  </DropdownMenu.Item>
{/if}

{#if allowCustomTimeRange}
  <DropdownMenu.Separator />
  <DropdownMenu.Item onclick={onSelectCustomOption} data-range="custom">
    <span class:font-bold={selected === "CUSTOM"}> {m.time_custom()} </span>
  </DropdownMenu.Item>
{/if}
