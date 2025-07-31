<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { DateTime, Duration, type DateTimeUnit } from "luxon";
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
  import { onDestroy, onMount } from "svelte";
  import SyntaxElement from "../components/SyntaxElement.svelte";

  export let dateTimeAnchor: DateTime;
  export let grain: V1TimeGrain | undefined;
  export let rangeGrain: V1TimeGrain | undefined;
  export let smallestTimeGrain: V1TimeGrain | undefined;
  export let snapToEnd: boolean;
  export let isPeriodToDate: boolean;
  export let watermark: DateTime | undefined;
  export let latest: DateTime | undefined;
  export let zone: string;
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
  let now = DateTime.now().setZone(zone);
  let interval: ReturnType<typeof setInterval> | undefined = undefined;

  onMount(() => {
    interval = setInterval(() => {
      now = DateTime.now().setZone(zone);
    }, 1000);
  });

  onDestroy(() => {
    if (interval) {
      clearInterval(interval);
    }
  });

  $: dateTimeUnit = grain ? V1TimeGrainToDateTimeUnit[grain] : undefined;

  $: grainOptions = getOptionsFromSmallestToLargest(
    rangeGrain,
    smallestTimeGrain,
    isPeriodToDate,
  );

  $: humanizedRef = humanizeRef(ref, grain);

  $: derivedAnchor = deriveAnchor(dateTimeAnchor, dateTimeUnit, snapToEnd);

  $: options = [
    {
      id: "watermark",
      label: "complete data",
      timestamp: watermark,
      description:
        "Timestamp prior to which data frames are considered complete, also known as the watermark",
    },
    {
      id: "latest",
      label: "latest data",
      timestamp: latest,
      description: "Timestamp of latest data point",
    },
    {
      id: "now",
      label: "current time",
      timestamp: now,
      description: "Server clock in selected timezone",
    },
  ];

  function deriveAnchor(
    dateTimeAnchor: DateTime,
    snap: DateTimeUnit | undefined,
    inclusive: boolean,
  ) {
    if (!snap) {
      return dateTimeAnchor;
    }
    return dateTimeAnchor.startOf(snap).plus({
      [snap]: inclusive ? 1 : 0,
    });
  }

  function humanizeRef(ref: string, grain: V1TimeGrain | undefined): string {
    switch (ref) {
      case "watermark":
        if (grain) return "complete";
        return "complete data";
      case "latest":
        return "latest";
      case "now":
        if (grain) return "current";
        return "now";
      default:
        return ref;
    }
  }

  function getColloquialOffset(date: DateTime): string {
    const inFuture = date > DateTime.now();
    return (
      Duration.fromObject(
        Object.fromEntries(
          Object.entries(
            DateTime.now().setZone(date.zone).diff(date).rescale().toObject(),
          )
            .filter(([, value]) => value !== 0)
            .slice(0, 2),
        ),
      ).toHuman({
        listStyle: "narrow",
        maximumFractionDigits: 0,
        signDisplay: "never",
      }) + (inFuture ? " from now" : " ago")
    );
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
              {#if snapToEnd || ref === "watermark"}
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

      <Tooltip.Content side="bottom" sideOffset={8} class="z-50">
        <TooltipContent>
          <TooltipTitle>
            <svelte:fragment slot="name">
              {derivedAnchor.toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS)}
            </svelte:fragment>
          </TooltipTitle>
          <TooltipDescription>
            {getColloquialOffset(derivedAnchor)}
          </TooltipDescription>
        </TooltipContent>
      </Tooltip.Content>
    </Tooltip.Root>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start" class="w-52 flex flex-col p-0">
    <DropdownMenu.Group class="p-1">
      <h3 class="mt-1 px-2 uppercase text-gray-500 font-semibold">Reference</h3>
      {#each options as { id, label, description, timestamp } (id)}
        {#if id !== "watermark" || (id === "watermark" && !!timestamp)}
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
                  <TooltipContent class="w-60">
                    <div class="flex items-center justify-between">
                      <span
                        class="font-bold truncate text-gray-100 dark:text-gray-200"
                      >
                        {timestamp.toLocaleString(
                          DateTime.DATETIME_MED_WITH_SECONDS,
                        )}
                      </span>
                      <SyntaxElement range={id} dark />
                    </div>

                    {#if id !== "now"}
                      <div>
                        {getColloquialOffset(timestamp)}
                      </div>
                    {/if}
                    <TooltipDescription>
                      {description}
                    </TooltipDescription>
                  </TooltipContent>
                </Tooltip.Content>
              {/if}
            </Tooltip.Root>
          </DropdownMenu.CheckboxItem>
        {/if}
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
          <span>Anchor to period end</span>

          <Switch
            disabled={ref === "watermark"}
            small
            checked={snapToEnd || ref === "watermark"}
            on:click={() => {
              onToggleAlignment(!snapToEnd);
            }}
          />
        </div>
      </div>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
