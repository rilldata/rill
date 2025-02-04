<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import type { V1ExploreTimeRange } from "@rilldata/web-common/runtime-client";
  import SyntaxElement from "./SyntaxElement.svelte";

  import {
    LATEST_WINDOW_TIME_RANGES,
    PERIOD_TO_DATE_RANGES,
    PREVIOUS_COMPLETE_DATE_RANGES,
  } from "@rilldata/web-common/lib/time/config";

  export let range: V1ExploreTimeRange;
  export let selected: boolean;
  export let onClick: ((range: string, syntax: boolean) => void) | undefined =
    undefined;

  $: meta = range.range?.startsWith("P")
    ? LATEST_WINDOW_TIME_RANGES[range.range]
    : range.range?.startsWith("rill")
      ? (PERIOD_TO_DATE_RANGES[range.range] ??
        PREVIOUS_COMPLETE_DATE_RANGES[range.range])
      : undefined;

  $: label = getRangeLabel(range.range ?? "");
</script>

<DropdownMenu.Item
  on:click={() => {
    console.log(!meta && !range.range?.startsWith("P"));
    if (onClick)
      onClick(
        meta?.rillSyntax ?? range.range,
        !meta && !range.range?.startsWith("P"),
      );
  }}
  class="group h-7"
>
  <div class="size-full flex justify-between items-center" title={range.range}>
    <span class:font-bold={selected} class="truncate">
      {#if label.endsWith(", complete")}
        {label.replace(", complete", "")}
        <span class="text-gray-400 text-[11px]">(complete)</span>
      {:else}
        {label}
      {/if}
    </span>

    {#if meta}
      <SyntaxElement range={meta?.rillSyntax ?? range.range} />
    {/if}
  </div>
</DropdownMenu.Item>
