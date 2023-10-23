<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipDescription from "@rilldata/web-common/components/tooltip/TooltipDescription.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { crossfade, fly } from "svelte/transition";
  import Spinner from "../../entity-management/Spinner.svelte";
  import {
    formatMeasurePercentageDifference,
    humanizeDataType,
    FormatPreset,
    humanizeDataTypeExpanded,
  } from "../humanize-numbers";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

  export let measure: MetricsViewSpecMeasureV2;
  export let value: number;
  export let comparisonOption: TimeComparisonOption = undefined;
  export let comparisonValue: number = undefined;
  export let comparisonPercChange: number = undefined;
  export let showComparison = false;
  export let status: EntityStatus;
  export let withTimeseries = true;
  export let isMeasureExpanded = false;

  const dispatch = createEventDispatcher();

  $: description =
    measure?.description || measure?.label || measure?.expression;
  $: formatPreset =
    (measure?.formatPreset as FormatPreset) || FormatPreset.HUMANIZE;

  $: name = measure?.label || measure?.expression;

  $: valueIsPresent = value !== undefined && value !== null;

  $: isComparisonPositive = Number.isFinite(diff) && (diff as number) >= 0;
  const [send, receive] = crossfade({ fallback: fly });

  $: diff = comparisonValue ? value - comparisonValue : false;
  $: noChange = !diff;

  $: formattedDiff = `${isComparisonPositive ? "+" : ""}${humanizeDataType(
    diff,
    formatPreset
  )}`;

  /** when the measure is a percentage, we don't show a percentage change. */
  $: measureIsPercentage = formatPreset === FormatPreset.PERCENTAGE;
</script>

<button
  on:click={() => dispatch("expand-measure")}
  class="big-number flex flex-col px-2 text-left rounded
  {isMeasureExpanded ? 'pointer-events-none' : 'hover:bg-gray-100'}
  {withTimeseries ? 'py-3' : 'py-1 justify-between'}"
>
  <Tooltip distance={16} location="top" alignment="start">
    <h2
      class="break-words line-clamp-2"
      style:font-size={withTimeseries ? "" : "0.8rem"}
    >
      {name}
    </h2>
    <TooltipContent slot="tooltip-content" maxWidth="280px">
      <TooltipDescription>
        {description}
      </TooltipDescription>
    </TooltipContent>
  </Tooltip>
  <div
    class="ui-copy-muted relative"
    style:font-size={withTimeseries ? "1.6rem" : "1.8rem"}
    style:font-weight="light"
  >
    <div>
      {#if valueIsPresent && status === EntityStatus.Idle}
        <Tooltip distance={8} location="bottom" alignment="start">
          <div class="w-max">
            <WithTween {value} tweenProps={{ duration: 500 }} let:output>
              {humanizeDataType(output, formatPreset)}
            </WithTween>
          </div>
          <TooltipContent slot="tooltip-content">
            {humanizeDataTypeExpanded(value, formatPreset)}
            <TooltipDescription>
              the aggregate value over the current time period
            </TooltipDescription>
          </TooltipContent>
        </Tooltip>
        {#if showComparison}
          <Tooltip distance={8} location="bottom" alignment="start">
            <div class="flex items-baseline gap-x-3">
              {#if comparisonValue != null}
                <div
                  class="w-max text-sm ui-copy-inactive"
                  class:font-semibold={isComparisonPositive}
                >
                  {#if !noChange}
                    {formattedDiff}
                  {:else}
                    <span
                      class="ui-copy-disabled-faint italic"
                      style:font-size=".9em">no change</span
                    >
                  {/if}
                </div>
              {/if}
              {#if comparisonPercChange != null && !noChange && !measureIsPercentage}
                <div
                  class="w-max text-sm
              {isComparisonPositive ? 'ui-copy-inactive' : 'text-red-500'}"
                >
                  <WithTween
                    value={comparisonPercChange}
                    tweenProps={{ duration: 500 }}
                    let:output
                  >
                    <PercentageChange
                      tabularNumber={false}
                      value={formatMeasurePercentageDifference(output)}
                    />
                  </WithTween>
                </div>
              {/if}
            </div>
            <TooltipContent slot="tooltip-content" maxWidth="300px">
              {@const tooltipPercentage =
                formatMeasurePercentageDifference(comparisonPercChange)}

              {#if noChange}
                no change over {TIME_COMPARISON[comparisonOption].shorthand}
              {:else}
                {TIME_COMPARISON[comparisonOption].shorthand}
                <span class="font-semibold">
                  {humanizeDataType(comparisonValue, formatPreset)}
                </span>
                {#if !measureIsPercentage}
                  <span class="text-gray-300">,</span>
                  <span
                    >{tooltipPercentage.int}% {isComparisonPositive
                      ? "increase"
                      : "decrease"}</span
                  >
                {/if}
              {/if}
            </TooltipContent>
          </Tooltip>
        {/if}
      {:else if status === EntityStatus.Error}
        <CrossIcon />
      {:else if status === EntityStatus.Running}
        <div
          class="{withTimeseries ? '' : 'bottom-0'} absolute p-2"
          in:receive|local={{ key: "spinner" }}
          out:send|local={{ key: "spinner" }}
        >
          <Spinner status={EntityStatus.Running} />
        </div>
      {:else if !valueIsPresent}
        <span class="ui-copy-disabled-faint italic text-sm">no data</span>
      {/if}
    </div>
  </div>
</button>

<style>
  .big-number:hover + .time-series-body {
    border-top: 1px solid var(--color-gray-200);
  }
</style>
