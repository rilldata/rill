<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { Interval } from "luxon";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { parseRillTime } from "../../../url-state/time-ranges/parser";
  import {
    getOptionsFromSmallestToLargest,
    GrainAliasToV1TimeGrain,
    V1TimeGrainToAlias,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import type {
    RillTime,
    RillTimeMeta,
  } from "../../../url-state/time-ranges/RillTime";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import DropdownMenuCheckboxItem from "@rilldata/web-common/components/dropdown-menu/DropdownMenuCheckboxItem.svelte";

  // export let timeString: string | undefined;
  export let parsedTime: RillTime;
  export let isComplete: boolean;
  export let inGrain: V1TimeGrain;
  export let rangeGrain: V1TimeGrain;
  // export let meta: RillTimeMeta;
  // export let interval: Interval<true>;
  // export let timeGrainOptions: V1TimeGrain[];
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let onSelectEnding: (
    grain: V1TimeGrain | undefined,
    complete?: boolean,
  ) => void;

  let open = false;

  $: complete = isComplete || smallestTimeGrain === inGrain;

  let prefersComplete = complete;

  // $: isComplete = parsedTime.isComplete;
  // $: inGrain = parsedTime.inGrain;
  // $: rangeGrain = parsedTime.rangeGrain;

  $: grainOptions = getOptionsFromSmallestToLargest(
    rangeGrain,
    smallestTimeGrain,
  );

  // let selected: "days" | "months" | "years" = "days";

  // $: alts = getAlts(meta, smallestTimeGrain);
  // $: selectedAlt = alts.find((a) => a.string === timeString);

  // function getAlts(
  //   meta: RillTimeMeta,
  //   smallestTimeGrain: V1TimeGrain | undefined,
  // ) {
  //   const integer = meta.integer;
  //   const primaryGrainAlias = meta.grain;

  //   if (integer === undefined || primaryGrainAlias === undefined) {
  //     return [];
  //   }

  //   const v1TimeGrain = GrainAliasToV1TimeGrain[primaryGrainAlias];
  //   const primaryGrainUnit = V1TimeGrainToDateTimeUnit[v1TimeGrain];
  //   const allowedGrains = getToDateExcludeOptions(
  //     v1TimeGrain,
  //     smallestTimeGrain,
  //   );

  //   if (meta.type === "lastN") {
  //     return [
  //       {
  //         string: `-${integer}${primaryGrainAlias}^ to ${primaryGrainAlias}^`,
  //         label: `excluding latest ${primaryGrainUnit}`,
  //       },
  //       {
  //         string: `-${integer}${primaryGrainAlias} to latest`,
  //         label: "to now",
  //       },
  //       {
  //         string: `-${integer}${primaryGrainAlias}$ to ${primaryGrainAlias}$`,
  //         label: `including latest ${primaryGrainUnit}`,
  //       },
  //     ];
  //   } else if (meta.type === "toDate") {
  //     return [
  //       {
  //         string: `${primaryGrainAlias}!`,
  //         label: parseRillTime(`${primaryGrainAlias}!`).getLabel(),
  //         alts: allowedGrains.map((g) => {
  //           if (g === v1TimeGrain) {
  //             return {
  //               string: `${primaryGrainAlias}^ to ${primaryGrainAlias}$`,
  //               label: `in full`,
  //             };
  //           }

  //           const grainAlias = V1TimeGrainToAlias[g];
  //           const unit = V1TimeGrainToDateTimeUnit[g];
  //           return {
  //             string: `${primaryGrainAlias}^ to ${grainAlias}^`,
  //             label: `excluding this ${unit}`,
  //           };
  //         }),
  //       },
  //     ];
  //   }

  //   return [];
  // }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      {...builder}
      use:builder.action
      class="flex gap-x-1 items-center"
      aria-label="Select time range"
      data-state={open ? "open" : "closed"}
    >
      {#if complete}
        <b>in complete</b>
      {:else}
        <b>in</b>
      {/if}
      <p>{V1TimeGrainToDateTimeUnit[inGrain]}s</p>

      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    align="start"
    class="w-fit overflow-hidden flex flex-col"
  >
    {#each grainOptions as option, i (i)}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={option === inGrain}
        on:click={() => {
          onSelectEnding(option, prefersComplete);
        }}
      >
        {V1TimeGrainToDateTimeUnit[option]}s
      </DropdownMenu.CheckboxItem>
    {/each}

    <div class="flex items-center gap-x-2 border-t p-2 pb-1 mt-1">
      Complete periods only <Switch
        disabled={smallestTimeGrain === inGrain}
        small
        on:click={() => {
          prefersComplete = !prefersComplete;
          onSelectEnding(inGrain, prefersComplete);
        }}
        checked={complete || prefersComplete}
      />
    </div>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style>
  /* The wrapper shrinks to the width of its content */
  .wrapper {
    display: inline-grid;
    grid-template-columns: 1fr; /* single column that both items share */
  }

  /* Vertical scroll container has an explicit width */
  .vertical-scroll {
    overflow-y: auto;
  }

  /* Horizontal container becomes a grid item and stretches to fill the column */
  .horizontal-scroll {
    overflow-x: auto;
    white-space: nowrap;

    /* No explicit width is set here */
  }
</style>
