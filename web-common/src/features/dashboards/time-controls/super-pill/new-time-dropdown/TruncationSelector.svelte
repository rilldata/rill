<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime, type DateTimeUnit } from "luxon";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import {
    getOptionsFromSmallestToLargest,
    V1TimeGrainToDateTimeUnit,
  } from "@rilldata/web-common/lib/time/new-grains";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { builderActions, getAttrs, Tooltip } from "bits-ui";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import TooltipDescription from "@rilldata/web-common/components/tooltip/TooltipDescription.svelte";

  export let dateTimeAnchor: DateTime;
  export let grain: V1TimeGrain | undefined;
  export let rangeGrain: V1TimeGrain | undefined;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let inclusive: boolean;
  export let watermark: DateTime | undefined;
  export let latest: DateTime | undefined;
  export let ref: "latest" | "watermark" | "now" | string;
  export let onSelectAsOfOption: (
    ref: "latest" | "watermark" | "now" | string,
  ) => void;
  export let onToggleAlignment: (forward: boolean) => void;
  export let onSelectEnding: (
    grain: V1TimeGrain | undefined,
    complete?: boolean,
  ) => void;

  let open = false;
  let now = DateTime.now();

  $: dateTimeUnit = grain ? V1TimeGrainToDateTimeUnit[grain] : undefined;

  $: grainOptions = getOptionsFromSmallestToLargest(
    rangeGrain,
    smallestTimeGrain,
  );

  $: humanizedRef = humanizeRef(ref, grain);

  $: derivedAnchor = deriveAnchor(dateTimeAnchor, dateTimeUnit, inclusive);

  $: options = [
    {
      id: "watermark",
      label: "complete data",
      timestamp: watermark,
      description: "Time prior to which data is considered complete",
    },
    {
      id: "latest",
      label: "latest data",
      timestamp: latest,
      description: "Timestamp of the latest data point",
    },
    {
      id: "now",
      label: "current time",
      timestamp: now,
      description: "System clock in the selected time zone",
    },
  ];

  function deriveAnchor(
    dateTimeAnchor: DateTime,
    snap: DateTimeUnit | undefined,
    inclusive: boolean,
  ) {
    if (!snap) {
      return dateTimeAnchor.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS);
    }
    return dateTimeAnchor
      .startOf(snap)
      .plus({
        [snap]: inclusive ? 1 : 0,
      })
      .toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS);
  }

  function humanizeRef(ref: string, grain: V1TimeGrain | undefined): string {
    switch (ref) {
      case "watermark":
        if (grain) return "complete";
        return "watermark";
      case "latest":
        return "latest";
      case "now":
        if (grain) return "current";
        return "now";
      default:
        return ref;
    }
  }
</script>

<DropdownMenu.Root bind:open disableFocusFirstItem={true}>
  <DropdownMenu.Trigger asChild let:builder id="truncation-selector-trigger">
    <Tooltip.Root openDelay={800}>
      <Tooltip.Trigger
        asChild
        let:builder={builder2}
        id="truncation-selector-trigger"
      >
        <button
          {...getAttrs([builder, builder2])}
          use:builderActions={{ builders: [builder, builder2] }}
          class="flex gap-x-1 items-center flex-none truncate"
          aria-label="Select time range"
          data-state={open ? "open" : "closed"}
        >
          <p>
            as of
            <b>
              {humanizedRef}
              {#if dateTimeUnit}
                {dateTimeUnit}
              {/if}
            </b>
            {#if grain}
              {#if inclusive || ref === "watermark"}
                end
              {:else}
                start
              {/if}
            {/if}
          </p>

          <span class="flex-none transition-transform" class:-rotate-180={open}>
            <CaretDownIcon />
          </span>
        </button>
      </Tooltip.Trigger>

      <Tooltip.Content side="bottom" sideOffset={8}>
        <TooltipContent>
          {derivedAnchor}
        </TooltipContent>
      </Tooltip.Content>
    </Tooltip.Root>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="w-52 flex flex-col p-0">
    <DropdownMenu.Group class="p-1">
      <h3 class="mt-1 px-2 uppercase text-gray-500 font-semibold">Reference</h3>
      {#each options as { id, label, description, timestamp } (id)}
        <DropdownMenu.CheckboxItem
          checkRight
          checked={ref === id}
          on:click={() => {
            onSelectAsOfOption(id);
          }}
        >
          <Tooltip.Root>
            <Tooltip.Trigger class="size-full flex justify-between ">
              {label}
            </Tooltip.Trigger>

            {#if timestamp}
              <Tooltip.Content side="right" sideOffset={40} class="w-65 z-50">
                <TooltipContent>
                  <TooltipTitle>
                    <svelte:fragment slot="name">
                      {timestamp.toLocaleString(
                        DateTime.DATETIME_MED_WITH_SECONDS,
                      )}
                    </svelte:fragment>
                  </TooltipTitle>
                  <TooltipDescription>
                    {description}
                  </TooltipDescription>
                </TooltipContent>
              </Tooltip.Content>
            {/if}
          </Tooltip.Root>
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Group>
    <DropdownMenu.Separator class="my-0" />

    <DropdownMenu.Group class="p-1">
      <h3 class="mt-1 px-2 uppercase text-gray-500 font-semibold">Grain</h3>

      {#each grainOptions as option, i (i)}
        <DropdownMenu.CheckboxItem
          checkRight
          checked={option === grain}
          on:click={() => {
            onSelectEnding(option);
          }}
        >
          {V1TimeGrainToDateTimeUnit[option]}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Group>

    {#if dateTimeUnit}
      <div class="bg-gray-100 border-t">
        <div class="flex justify-between items-center p-2">
          <span>
            Include

            {#if ref === "latest"}
              latest
            {:else if ref === "now"}
              current
            {:else if ref === "watermark"}
              last complete
            {/if}

            {dateTimeUnit}
          </span>

          <Switch
            disabled={ref === "watermark"}
            small
            checked={inclusive || ref === "watermark"}
            on:click={() => {
              onToggleAlignment(!inclusive);
            }}
          />
        </div>
      </div>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  h3 {
    @apply text-[11px] text-gray-500;
  }
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
