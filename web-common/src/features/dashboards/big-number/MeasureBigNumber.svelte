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
  import {
    type MetricsViewSpecMeasure,
    createQueryServiceMetricsViewAggregation,
    type V1Expression,
  } from "@rilldata/web-common/runtime-client";
  import { cellInspectorStore } from "../stores/cell-inspector-store";
  import {
    crossfade,
    fly,
    type CrossfadeParams,
    type FlyParams,
  } from "svelte/transition";
  import BigNumberTooltipContent from "./BigNumberTooltipContent.svelte";
  import { keepPreviousData, type QueryClient } from "@tanstack/svelte-query";

  export let measure: MetricsViewSpecMeasure;
  export let withTimeseries = true;
  export let isMeasureExpanded = false;

  // Query-context props
  export let instanceId: string;
  export let metricsViewName: string;
  export let where: V1Expression | undefined = undefined;
  export let timeDimension: string | undefined = undefined;
  export let timeStart: string | undefined = undefined;
  export let timeEnd: string | undefined = undefined;
  export let comparisonTimeStart: string | undefined = undefined;
  export let comparisonTimeEnd: string | undefined = undefined;
  export let showComparison = false;
  export let ready: boolean = true;

  $: measureName = measure.name ?? "";

  // Primary totals query
  $: primaryQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: measureName }],
      where,
      timeRange: {
        start: timeStart,
        end: timeEnd,
        timeDimension,
      },
    },
    {
      query: {
        enabled: ready && !!timeStart && !!measureName,
        placeholderData: keepPreviousData,
        refetchOnMount: false,
      },
    },
  );

  // Comparison totals query
  $: comparisonQuery = createQueryServiceMetricsViewAggregation(
    instanceId,
    metricsViewName,
    {
      measures: [{ name: measureName }],
      where,
      timeRange: {
        start: comparisonTimeStart,
        end: comparisonTimeEnd,
        timeDimension,
      },
    },
    {
      query: {
        enabled:
          ready && showComparison && !!comparisonTimeStart && !!measureName,
        placeholderData: keepPreviousData,
        refetchOnMount: false,
      },
    },
  );

  // Derive value, comparisonValue, status, errorMessage from queries
  $: value =
    ($primaryQuery.data?.data?.[0]?.[measureName] as number | null) ?? null;
  $: comparisonValue = showComparison
    ? ($comparisonQuery.data?.data?.[0]?.[measureName] as number | undefined)
    : undefined;

  $: isFetching =
    $primaryQuery.isFetching || (showComparison && $comparisonQuery.isFetching);
  $: isError = $primaryQuery.isError || $comparisonQuery.isError;

  $: status = isError
    ? EntityStatus.Error
    : isFetching
      ? EntityStatus.Running
      : EntityStatus.Idle;

  $: errorMessage = isError
    ? (($primaryQuery.error as any)?.response?.data?.message ??
      ($comparisonQuery.error as any)?.response?.data?.message ??
      undefined)
    : undefined;

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

  function handleMouseOver() {
    if (value !== undefined && value !== null) {
      // Always update the value in the store, but don't change visibility
      cellInspectorStore.updateValue(value.toString());
    }
  }

  function handleFocus() {
    if (value !== undefined && value !== null) {
      // Always update the value in the store, but don't change visibility
      cellInspectorStore.updateValue(value.toString());
    }
  }
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
    value={tooltipValue}
  />

  <svelte:element
    this={useDiv ? "div" : "a"}
    role={useDiv ? "presentation" : "button"}
    tabindex={useDiv ? -1 : 0}
    class="group big-number outline-border"
    class:shadow-grad={!useDiv}
    class:cursor-pointer={!useDiv}
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
    href={tddHref}
  >
    <h2
      class="line-clamp-2 text-fg-muted hover:text-theme-700 group-hover:text-theme-700 font-semibold whitespace-normal"
      style:font-size={withTimeseries ? "" : "0.8rem"}
    >
      {name}
    </h2>
    <div
      role="button"
      class="text-fg-secondary relative w-full h-full overflow-hidden text-ellipsis"
      style:font-size={withTimeseries ? "1.6rem" : "1.8rem"}
      style:font-weight="light"
      on:mouseover={handleMouseOver}
      on:focus={handleFocus}
      tabindex="0"
    >
      {#if value !== null && value !== undefined && status === EntityStatus.Idle}
        <WithTween {value} tweenProps={{ duration: 500 }} let:output>
          {measureValueFormatter(output)}
        </WithTween>
        {#if showComparison && comparisonValue}
          <div class="flex items-baseline gap-x-3 text-sm">
            {#if comparisonValue != null}
              <div
                role="complementary"
                class="w-fit max-w-full overflow-hidden text-ellipsis text-fg-secondary"
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
                  <span class="text-fg-muted italic" style:font-size=".9em"
                    >no change</span
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
                class="w-fit text-fg-secondary"
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
        <span class="text-fg-muted italic text-sm">no data</span>
      {:else if value === undefined}
        <span class="text-fg-muted italic text-sm">n/a</span>
      {/if}
    </div>
  </svelte:element>
</Tooltip>

<style lang="postcss">
  .big-number {
    @apply h-fit w-[138px] m-0.5 rounded p-2 font-normal;
    @apply items-start flex flex-col text-left flex-none;
    min-height: 85px;
  }

  .shadow-grad:hover {
    @apply shadow-md outline-1 outline;
    outline-color: color-mix(
      in oklab,
      var(--color-theme-500) calc(0.15 * 100%),
      transparent
    );

    background: linear-gradient(
      to bottom,
      color-mix(in oklab, var(--white) calc(0.15 * 100%), transparent),
      50%,
      color-mix(in oklab, var(--color-theme-300) calc(0.1 * 100%), transparent)
    );
  }

  :global(.dark) .shadow-grad:hover {
    @apply shadow-md  outline-1 outline outline-[#FFFFFF26];
    background: linear-gradient(
      to bottom,
      color-mix(in oklab, var(--white) calc(0.1 * 100%), transparent),
      50%,
      color-mix(in oklab, var(--white) calc(0.05 * 100%), transparent)
    );
  }
</style>
