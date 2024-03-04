<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import type {
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import {
    CrossfadeParams,
    FlyParams,
    crossfade,
    fly,
  } from "svelte/transition";
  import Spinner from "../../entity-management/Spinner.svelte";
  import BigNumberTooltipContent from "./BigNumberTooltipContent.svelte";

  export let measure: MetricsViewSpecMeasureV2;
  export let value: number | null;
  export let comparisonOption:
    | TimeComparisonOption
    | TimeRangePreset
    | undefined = undefined;
  export let comparisonValue: number | undefined = undefined;
  export let showComparison = false;
  export let status: EntityStatus;
  export let withTimeseries = true;
  export let isMeasureExpanded = false;

  $: comparisonPercChange =
    comparisonValue && value !== undefined && value !== null
      ? (value - comparisonValue) / comparisonValue
      : undefined;

  const dispatch = createEventDispatcher();

  $: measureValueFormatter = createMeasureValueFormatter<null>(measure);

  // this is used to show the full value in tooltips when the user hovers
  // over the number. If not present, we'll use the string "no data"
  $: measureValueFormatterUnabridged = createMeasureValueFormatter<null>(
    measure,
    true,
  );

  $: name = measure?.label || measure?.expression;

  $: if (value === undefined) {
    value = null;
  }

  const [send, receive] = crossfade({
    fallback: (node: Element, params: CrossfadeParams) =>
      fly(node, params as FlyParams),
  });

  $: diff =
    value !== null && comparisonValue !== undefined
      ? value - comparisonValue
      : 0;
  $: noChange = !comparisonValue;
  $: isComparisonPositive = diff >= 0;

  $: formattedDiff = `${isComparisonPositive ? "+" : ""}${measureValueFormatter(
    diff,
  )}`;

  /** when the measure is a percentage, we don't show a percentage change. */
  $: measureIsPercentage = measure?.formatPreset === FormatPreset.PERCENTAGE;

  $: hoveredValue = measureValueFormatterUnabridged(value) ?? "no data";

  const { shiftClickAction } = createShiftClickAction();
  async function shiftClickHandler(number: string | undefined) {
    if (number === undefined) return;
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

  <svelte:element
    this={isMeasureExpanded ? "div" : "button"}
    role={isMeasureExpanded ? "presentation" : "button"}
    tabindex={isMeasureExpanded ? -1 : 0}
    class="group big-number m-0.5 rounded flex items-start"
    class:shadow-grad={!isMeasureExpanded}
    class:cursor-pointer={!isMeasureExpanded}
    on:click={(e) => {
      if (e.shiftKey) return;
      dispatch("expand-measure");
    }}
    on:shift-click={() => shiftClickHandler(hoveredValue)}
    use:shiftClickAction
  >
    <div
      class="flex flex-col px-2 text-left w-full h-full
    {withTimeseries ? 'py-3' : 'py-1 justify-between'}
    "
    >
      <h2
        style:overflow-wrap="anywhere"
        class="line-clamp-2 ui-header-primary font-semibold"
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
          {#if value !== null && status === EntityStatus.Idle}
            <div class="w-full overflow-hidden text-ellipsis">
              <WithTween {value} tweenProps={{ duration: 500 }} let:output>
                {measureValueFormatter(output)}
              </WithTween>
            </div>
            {#if showComparison && comparisonOption && comparisonValue}
              <div class="flex items-baseline gap-x-3">
                {#if comparisonValue != null}
                  <div
                    role="complementary"
                    on:mouseenter={() =>
                      (hoveredValue =
                        measureValueFormatterUnabridged(diff) ?? "no data")}
                    on:mouseleave={() =>
                      (hoveredValue =
                        measureValueFormatterUnabridged(value) ?? "no data")}
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
                    role="complementary"
                    on:mouseenter={() =>
                      (hoveredValue = numberPartsToString(
                        formatMeasurePercentageDifference(
                          comparisonPercChange ?? 0,
                        ),
                      ))}
                    on:mouseleave={() =>
                      (hoveredValue =
                        measureValueFormatterUnabridged(value) ?? "no data")}
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
              class="absolute p-2"
              class:bottom-0={withTimeseries}
              in:receive={{ key: "spinner" }}
              out:send={{ key: "spinner" }}
            >
              <Spinner status={EntityStatus.Running} />
            </div>
          {:else if value === null}
            <span class="ui-copy-disabled-faint italic text-sm">no data</span>
          {/if}
        </div>
      </div>
    </div>
  </svelte:element>
</Tooltip>

<style>
  .big-number {
    width: 118px;
    height: 85px;
  }

  .shadow-grad:hover {
    /* ui-card */
    background: var(
      --gradient_white-slate50,
      linear-gradient(180deg, #fff 0%, #f8fafc 100%)
    );
    box-shadow:
      0px 4px 6px 0px rgba(15, 23, 42, 0.09),
      0px 0px 0px 1px rgba(15, 23, 42, 0.06),
      0px 1px 3px 0px rgba(15, 23, 42, 0.04),
      0px 2px 3px 0px rgba(15, 23, 42, 0.03);
  }
</style>
