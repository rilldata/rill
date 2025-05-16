<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
  import SyntaxElement from "./SyntaxElement.svelte";
  import type { UITimeRange } from "../../time-range-store";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import {
    getGrainAliasFromString,
    getLowerOrderGrain,
    getNextLowerOrderDuration,
  } from "@rilldata/web-common/lib/time/new-grains";
  import type { RillTime } from "../../../url-state/time-ranges/RillTime";

  export let range: string;
  export let parsed: RillTime;
  export let type: "this" | "last" | "ago" | "by";
  export let selected: boolean;
  // export let smallestTimeGrain: V1TimeGrain | undefined;
  export let onClick: (range: string, syntax: boolean) => void;

  // $: disabled = range.parsed?.timeRangeGrain;

  $: label = parsed?.getLabel();

  // $: finalRange = range?.meta?.rillSyntax ?? range.range;

  // $: isRillSyntax = !(
  //   finalRange?.startsWith("P") || finalRange?.startsWith("rill")
  // );
</script>

<div
  role="presentation"
  on:click={() => {
    onClick(range, true);
  }}
  class="group h-7 pr-2 overflow-hidden hover:bg-gray-100 rounded-sm w-full select-none flex items-center"
>
  <button
    class:font-bold={selected}
    class="truncate w-full text-left flex-shrink pl-2 h-full"
  >
    {#if label.endsWith(", complete")}
      {label.replace(", complete", "")}
      <!-- <span class="text-gray-400 text-[11px]">(complete)</span> -->
    {:else}
      {label}
    {/if}
  </button>

  <div class="flex gap-x-2 overflow-hidden flex-none">
    {#if type === "last" || type === "by"}
      <button
        class="sub"
        on:click={() => {
          // remove possible tilde from end
          const rangeWithoutTilde = range.replace(/~$/, "");

          onClick(rangeWithoutTilde, true);
        }}
      >
        complete
      </button>
    {/if}

    {#if type === "last"}
      <button
        class="sub"
        on:click={() => {
          const nextRange = getNextLowerOrderDuration(range);

          if (nextRange) onClick(nextRange, true);
        }}
      >
        to date
      </button>
    {/if}

    {#if type === "this"}
      <button
        class="sub"
        on:click={() => {
          const grainAlias = getGrainAliasFromString(range);
          console.log("grainAlias", grainAlias);
        }}
      >
        so far
      </button>
    {/if}
  </div>
</div>

<style lang="postcss">
  .sub {
    @apply px-2 bg-surface rounded-full border font-medium  border-gray-200 flex-none group-hover:opacity-100 opacity-0;
  }

  .sub:hover {
    @apply border-gray-300 bg-gray-50;
  }
</style>
