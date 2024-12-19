<script lang="ts">
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { numberPartsToString } from "@rilldata/web-common/lib/number-formatting/utils/number-parts-utils";
  import { type MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import {
    crossfade,
    fly,
    type CrossfadeParams,
    type FlyParams,
  } from "svelte/transition";
  import BigNumberTooltipContent from "./BigNumberTooltipContent.svelte";

  export let measure: MetricsViewSpecMeasureV2;
  export let value: number | null;

  export let comparisonValue: number | undefined = undefined;
  export let showComparison = false;
  export let status: EntityStatus;
  export let errorMessage: string | undefined = undefined;
  export let withTimeseries = true;
  export let isMeasureExpanded = false;

  $: comparisonPercChange =
    comparisonValue && value !== undefined && value !== null
      ? (value - comparisonValue) / comparisonValue
      : undefined;

  $: measureValueFormatter = createMeasureValueFormatter<null>(
    measure,
    "big-number",
  );

  // this is used to show the full value in tooltips when the user hovers
  // over the number. If not present, we'll use the string "no data"
  $: measureValueFormatterTooltip = createMeasureValueFormatter<null>(
    measure,
    "tooltip",
  );

  $: measureValueFormatterUnabridged = createMeasureValueFormatter<null>(
    measure,
    "unabridged",
  );

  $: name = measure?.displayName || measure?.expression;

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

  $: copyValue = measureValueFormatterUnabridged(value) ?? "no data";
  $: tooltipValue = measureValueFormatterTooltip(value) ?? "no data";

  $: tddHref = `?${ExploreStateURLParams.WebView}=tdd&${ExploreStateURLParams.ExpandedMeasure}=${measure.name}`;

  function shiftClickHandler(number: string | undefined) {
    if (number === undefined) return;

    copyToClipboard(number, `copied measure value "${number}" to clipboard`);
  }

  let suppressTooltip = false;

  const handleExpandMeasure = () => {
    if (!isMeasureExpanded) {
      isMeasureExpanded = true;
    }
  };
  $: useDiv = isMeasureExpanded || !withTimeseries;
</script>

<Tooltip
  alignment="start"
  distance={8}
  location="right"
  suppress={suppressTooltip}
>
  <BigNumberTooltipContent
    isMeasureExpanded={useDiv}
    {measure}
    slot="tooltip-content"
    value={tooltipValue}
  />

  <svelte:element
    class="group big-number"
    class:cursor-pointer={!useDiv}
    class:shadow-grad={!useDiv}
    href={tddHref}
    on:click={modified({
      shift: () => shiftClickHandler(copyValue),
      click: () => {
        suppressTooltip = true;
        handleExpandMeasure();
        setTimeout(() => {
          suppressTooltip = false;
        }, 1000);
      },
    })}
    role={useDiv ? "presentation" : "button"}
    tabindex={useDiv ? -1 : 0}
    this={useDiv ? "div" : "a"}
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
        {#if showComparison && comparisonValue}
          <div class="flex items-baseline gap-x-3 text-sm">
            {#if comparisonValue != null}
              <div
                role="complementary"
                class="w-fit max-w-full overflow-hidden text-ellipsis ui-copy-inactive"
                class:font-semibold={isComparisonPositive}
                on:mouseenter={() =>
                  (tooltipValue =
                    measureValueFormatterTooltip(diff) ?? "no data")}
                on:mouseleave={() =>
                  (tooltipValue =
                    measureValueFormatterTooltip(value) ?? "no data")}
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
                  (tooltipValue = numberPartsToString(
                    formatMeasurePercentageDifference(
                      comparisonPercChange ?? 0,
                    ),
                  ))}
                on:mouseleave={() =>
                  (tooltipValue =
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
        <div class="text-xs pt-1">
          {#if errorMessage}
            Error: {errorMessage}
          {:else}
            Error fetching totals data
          {/if}
        </div>
      {:else if status === EntityStatus.Running}
        <div
          class="absolute p-2"
          class:bottom-0={withTimeseries}
          in:receive={{ key: "spinner" }}
          out:send={{ key: "spinner" }}
        >
          <DelayedSpinner
            isLoading={status === EntityStatus.Running}
            size="24px"
          />
        </div>
      {:else if value === null}
        <span class="ui-copy-disabled-faint italic text-sm">no data</span>
      {/if}
    </div>
  </svelte:element>
</Tooltip>

<style lang="postcss">
  .big-number {
    @apply h-fit w-[138px] m-0.5 rounded p-2 font-normal;
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
