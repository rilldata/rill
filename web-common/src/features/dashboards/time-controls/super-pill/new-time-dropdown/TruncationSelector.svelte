<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime } from "luxon";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import {
    getOptionsFromSmallestToLargest,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";

  export let dateTimeAnchor: DateTime;
  export let grain: V1TimeGrain;
  export let rangeGrain: V1TimeGrain;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let inclusive: boolean;
  export let ref: "latest" | "watermark" | "now" | string;

  export let onToggleAlignment: (forward: boolean) => void;
  export let onSelectEnding: (
    grain: V1TimeGrain | undefined,
    complete?: boolean,
  ) => void;

  let open = false;

  $: grainOptions = getOptionsFromSmallestToLargest(
    rangeGrain,
    smallestTimeGrain,
  );
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
      <p>
        to

        <b>
          {V1TimeGrainToDateTimeUnit[grain]}
        </b>

        {#if inclusive}
          end
        {:else}
          start
        {/if}
      </p>

      <span class="flex-none transition-transform" class:-rotate-180={open}>
        <CaretDownIcon />
      </span>
    </button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    align="start"
    class="w-48 overflow-hidden flex flex-col p-0"
  >
    <DropdownMenu.Group class="p-1">
      {#each grainOptions as option, i (i)}
        <Tooltip alignment="end" location="right" distance={8}>
          <DropdownMenu.CheckboxItem
            checkRight
            checked={option === grain}
            on:click={() => {
              onSelectEnding(option);
            }}
          >
            {V1TimeGrainToDateTimeUnit[option]}
          </DropdownMenu.CheckboxItem>
          <TooltipContent slot="tooltip-content" maxWidth="600px">
            {dateTimeAnchor
              .startOf(V1TimeGrainToDateTimeUnit[option])
              .plus({ [V1TimeGrainToDateTimeUnit[option]]: inclusive ? 1 : 0 })
              .toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
          </TooltipContent>
        </Tooltip>
      {/each}
    </DropdownMenu.Group>

    {#if ref !== "watermark"}
      <div
        class="flex justify-between items-center p-1.5 px-3 bg-gray-100 border-t"
      >
        <span>
          Include

          {#if ref === "latest"}
            latest
          {:else if ref === "now"}
            current
          {/if}

          {V1TimeGrainToDateTimeUnit[grain]}
        </span>

        <Switch
          small
          checked={inclusive}
          on:click={() => {
            onToggleAlignment(!inclusive);
          }}
        />
      </div>
    {/if}
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
