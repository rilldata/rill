<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import CrossIcon from "@rilldata/web-common/components/icons/CrossIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { createShiftClickAction } from "@rilldata/web-common/lib/actions/shift-click-action";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
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
    eventBus.emit("notification", {
      message: `copied dimension value "${number}" to clipboard`,
    });
  }

  let suppressTooltip = false;
</script>

<Tooltip
  suppress={suppressTooltip}
  distance={8}
  location="right"
  alignment="start"
>
  <BigNumberTooltipContent
    slot="tooltip-content"
    {measure}
    value={hoveredValue}
  />

  <svelte:element
    this={isMeasureExpanded ? "div" : "button"}
    role={isMeasureExpanded ? "presentation" : "button"}
    tabindex={isMeasureExpanded ? -1 : 0}
    class="group big-number"
    class:shadow-grad={!isMeasureExpanded}
    class:cursor-pointer={!isMeasureExpanded}
    on:click={(e) => {
      if (e.shiftKey) return;
      suppressTooltip = true;
      dispatch("expand-measure");
      setTimeout(() => {
        suppressTooltip = false;
      }, 1000);
    }}
    on:shift-click={() => shiftClickHandler(hoveredValue)}
    use:shiftClickAction
  >
    <h2
      class="line-clamp-2 ui-header-primary font-semibold whitespace-normal"
      style:font-size={withTimeseries ? "" : "0.8rem"}
    >
      {name}
    </h2>
    <div
      class="ui-copy-muted relative w-full h-full overflow-hidden text-ellipsis"
      style:font-size={withTimeseries ? "1.6rem" : "1.8rem"}
      style:font-weight="light"
    >
      {#if value !== null && status === EntityStatus.Idle}
        <WithTween {value} tweenProps={{ duration: 500 }} let:output>
          {measureValueFormatter(output)}
        </WithTween>
        {#if showComparison && comparisonOption && comparisonValue}
          <div class="flex items-baseline gap-x-3 text-sm">
            {#if comparisonValue != null}
              <div
                role="complementary"
                class="w-fit max-w-full overflow-hidden text-ellipsis ui-copy-inactive"
                class:font-semibold={isComparisonPositive}
                on:mouseenter={() =>
                  (hoveredValue =
                    measureValueFormatterUnabridged(diff) ?? "no data")}
                on:mouseleave={() =>
                  (hoveredValue =
                    measureValueFormatterUnabridged(value) ?? "no data")}
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
                class="w-fit ui-copy-inactive"
                class:text-red-500={!isComparisonPositive}
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
  </svelte:element>
</Tooltip>

<style lang="postcss">
  .big-number {
    @apply h-fit w-[138px] m-0.5 rounded p-2;
    min-height: 85px;
    @apply items-start flex flex-col text-left;
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
