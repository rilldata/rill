<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import { crossfade, fly } from "svelte/transition";
  import Spinner from "../../entity-management/Spinner.svelte";
  import {
    formatMeasurePercentageDifference,
    humanizeDataType,
    NicelyFormattedTypes,
  } from "../humanize-numbers";

  export let value: number;
  export let comparisonOption: TimeComparisonOption;
  export let comparisonValue: number;
  export let comparisonPercChange: number;
  export let showComparison: boolean;

  export let status: EntityStatus;
  export let description: string = undefined;
  export let withTimeseries = true;
  export let formatPreset: string; // workaround, since unable to cast `string` to `NicelyFormattedTypes` within MetricsTimeSeriesCharts.svelte's `#each` block

  $: formatPresetEnum =
    (formatPreset as NicelyFormattedTypes) || NicelyFormattedTypes.HUMANIZE;
  $: valusIsPresent = value !== undefined && value !== null;

  $: isComparisonPositive = comparisonPercChange && comparisonPercChange >= 0;

  const [send, receive] = crossfade({ fallback: fly });

  $: diff = comparisonValue ? value - comparisonValue : false;
  $: noChange = !diff;
</script>

<div class="flex flex-col {withTimeseries ? 'mt-2' : 'justify-between'}">
  <Tooltip distance={16} location="top">
    <h2
      class="break-words line-clamp-2"
      style:font-size={withTimeseries ? "" : "0.8rem"}
    >
      <slot name="name" />
    </h2>
    <TooltipContent slot="tooltip-content">
      {description}
    </TooltipContent>
  </Tooltip>
  <div
    class="ui-copy-muted relative"
    style:font-size={withTimeseries ? "1.6rem" : "1.8rem"}
    style:font-weight="light"
  >
    <!-- the default slot will be a tweened number that uses the formatter. One can optionally
    override this by filling the slot in the consuming component. -->
    <slot name="value">
      <div>
        {#if valusIsPresent && status === EntityStatus.Idle}
          <Tooltip distance={8} location="bottom" alignment="start">
            <div class="w-max">
              <WithTween {value} tweenProps={{ duration: 500 }} let:output>
                {#if formatPresetEnum !== NicelyFormattedTypes.NONE}
                  {humanizeDataType(output, formatPresetEnum)}
                {:else}
                  {output}
                {/if}
              </WithTween>
            </div>
            <TooltipContent slot="tooltip-content">
              the aggregate value over the current time period
            </TooltipContent>
          </Tooltip>
          {#if showComparison}
            <Tooltip distance={8} location="bottom" alignment="start">
              <div class="flex items-baseline gap-x-3">
                {#if comparisonValue != null}
                  <div
                    class="w-max text-sm ui-copy-inactive "
                    class:font-semibold={isComparisonPositive}
                  >
                    <WithTween
                      value={comparisonValue}
                      tweenProps={{ duration: 500 }}
                      let:output
                    >
                      {@const formattedValue =
                        formatPresetEnum !== NicelyFormattedTypes.NONE
                          ? humanizeDataType(diff, formatPresetEnum)
                          : diff}
                      {#if !noChange}
                        {isComparisonPositive ? "+" : ""}{formattedValue}
                      {:else}
                        <span
                          class="ui-copy-disabled-faint italic"
                          style:font-size=".9em">no change</span
                        >
                      {/if}
                    </WithTween>
                  </div>
                {/if}
                {#if comparisonPercChange != null && !noChange}
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
                  <span class="font-semibold"
                    >{formatPresetEnum !== NicelyFormattedTypes.NONE
                      ? humanizeDataType(comparisonValue, formatPresetEnum)
                      : comparisonValue}</span
                  ><span class="text-gray-300">,</span>

                  <span
                    >{tooltipPercentage.int}% {isComparisonPositive
                      ? "increase"
                      : "decrease"}</span
                  >
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
        {/if}
      </div>
    </slot>
  </div>
</div>
