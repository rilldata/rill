<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
  import {
    CrossfadeParams,
    FlyParams,
    crossfade,
    fly,
  } from "svelte/transition";
  import Spinner from "../../entity-management/Spinner.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import BigNumberTooltipContent from "./BigNumberTooltipContent.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";

  export let measure: MetricsViewSpecMeasureV2;
  export let value: number;
  export let comparisonOption: TimeComparisonOption | undefined = undefined;
  export let comparisonValue: number | undefined = undefined;
  export let comparisonPercChange: number | undefined = undefined;
  export let showComparison = false;
  export let status: EntityStatus;
  export let withTimeseries = true;
  export let isMeasureExpanded = false;

  const dispatch = createEventDispatcher();

  $: measureValueFormatter = createMeasureValueFormatter(measure);

  $: measureValueFormatterUnabridged = createMeasureValueFormatter(
    measure,
    true
  );

  $: name = measure?.label || measure?.expression;

  $: valueIsPresent = value !== undefined && value !== null;

  const [send, receive] = crossfade({
    fallback: (node: Element, params: CrossfadeParams) =>
      fly(node, params as FlyParams),
  });

  $: diff =
    valueIsPresent && comparisonValue !== undefined
      ? value - comparisonValue
      : 0;
  $: noChange = !comparisonValue;
  $: isComparisonPositive = diff >= 0;

  $: formattedDiff = `${isComparisonPositive ? "+" : ""}${measureValueFormatter(
    diff
  )}`;

  /** when the measure is a percentage, we don't show a percentage change. */
  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: hoveredValue = measureValueFormatterUnabridged(value);

  const { shiftClickAction } = createShiftClickAction();
  async function shiftClickHandler(number: string) {
    await navigator.clipboard.writeText(number);
    notifications.send({
      message: `copied dimension value "${number}" to clipboard`,
    });
  }
</script>

<Tooltip distance={8} location="right" alignment="start">
  <BigNumberTooltipContent
    slot="tooltip-content"
    {measure}
    value={hoveredValue}
  />

  <div
    use:shiftClickAction
    on:shift-click={() => shiftClickHandler(hoveredValue)}
    class="big-number m-0.5 rounded"
  >
    <button
      on:click={(e) => {
        if (e.shiftKey) return;
        dispatch("expand-measure");
      }}
      class="flex flex-col px-2 text-left
    {withTimeseries ? 'py-3' : 'py-1 justify-between'}
    {isMeasureExpanded ? 'cursor-default' : ''}
    "
    >
      <h2
        style:overflow-wrap="anywhere"
        class="line-clamp-2"
        style:font-size={withTimeseries ? "" : "0.8rem"}
      >
        {name}
      </h2>
      <div
        class="ui-copy-muted relative"
        style:font-size={withTimeseries ? "1.6rem" : "1.8rem"}
        style:font-weight="light"
      >
        <div>
          {#if valueIsPresent && status === EntityStatus.Idle}
            <div class="w-max">
              <WithTween {value} tweenProps={{ duration: 500 }} let:output>
                {measureValueFormatter(output)}
              </WithTween>
            </div>
            {#if showComparison && comparisonOption && comparisonValue}
              <div class="flex items-baseline gap-x-3">
                {#if comparisonValue != null}
                  <div
                    on:mouseenter={() =>
                      (hoveredValue = measureValueFormatterUnabridged(diff))}
                    on:mouseleave={() =>
                      (hoveredValue = measureValueFormatterUnabridged(value))}
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
                    on:mouseenter={() =>
                      (hoveredValue = numberPartsToString(
                        formatMeasurePercentageDifference(
                          comparisonPercChange ?? 0
                        )
                      ))}
                    on:mouseleave={() =>
                      (hoveredValue = measureValueFormatterUnabridged(value))}
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
  </div>
</Tooltip>

<style lang="postcss">
  .big-number:hover {
    @apply ui-card;
  }
</style>
