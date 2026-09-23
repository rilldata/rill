<script lang="ts">
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type {
    RangeBuckets,
    NamedRange,
    ISODurationString,
    InheritRangeOption,
  } from "../../new-time-controls";
  import { RILL_TO_LABEL, ALL_TIME_RANGE_ALIAS } from "../../new-time-controls";

  export let ranges: RangeBuckets;
  export let selected: NamedRange | ISODurationString;
  export let showDefaultItem: boolean;
  export let inheritOption: InheritRangeOption | undefined = undefined;
  export let defaultTimeRange: NamedRange | ISODurationString | undefined;
  export let onSelectRange: (range: NamedRange | ISODurationString) => void;
  export let onSelectCustomOption: () => void;
  export let allowCustomTimeRange = true;

  // While the range is inherited none of the concrete ranges is the selection.
  $: highlighted = inheritOption?.selected ? undefined : selected;

  function handleClick(e: MouseEvent) {
    const range = (e.currentTarget as HTMLElement)?.dataset?.range;
    if (!range) {
      throw new Error("No range provided");
    }

    onSelectRange(range);
  }
</script>

{#if inheritOption}
  <DropdownMenu.Item onclick={inheritOption.onSelect}>
    <span class:font-bold={inheritOption.selected}>{inheritOption.label}</span>
  </DropdownMenu.Item>

  <DropdownMenu.Separator />
{/if}

{#if showDefaultItem && defaultTimeRange}
  <DropdownMenu.Item data-range={defaultTimeRange} onclick={handleClick}>
    <div class:font-bold={highlighted === defaultTimeRange}>
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
    <span class:font-bold={highlighted === rillTime.interval.toString()}>
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
    <span class:font-bold={highlighted === rillTime.interval.toString()}>
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
    <span class:font-bold={highlighted === rillTime.interval.toString()}>
      {rillTime.getLabel()}
    </span>
  </DropdownMenu.Item>
{/each}

{#if ranges.allTime}
  <DropdownMenu.Separator />
  <DropdownMenu.Item onclick={handleClick} data-range={ALL_TIME_RANGE_ALIAS}>
    <span class:font-bold={highlighted === ALL_TIME_RANGE_ALIAS}>
      {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
    </span>
  </DropdownMenu.Item>
{/if}

{#if allowCustomTimeRange}
  <DropdownMenu.Separator />
  <DropdownMenu.Item onclick={onSelectCustomOption} data-range="custom">
    <span class:font-bold={highlighted === "CUSTOM"}> {m.time_custom()} </span>
  </DropdownMenu.Item>
{/if}
