<script lang="ts">
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { slide } from "svelte/transition";
  import { type LeaderboardItemData, makeHref } from "./leaderboard-utils";
  import { useCellInspector } from "../hooks/useCellInspector";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    COMPARISON_COLUMN_WIDTH,
    DEFAULT_COLUMN_WIDTH,
    valueColumn,
    deltaColumn,
  } from "./leaderboard-widths";
  import FloatingElement from "@rilldata/web-common/components/floating-element/FloatingElement.svelte";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
  export let tableWidth: number;
  export let borderTop = false;
  export let borderBottom = false;
  export let isBeingCompared: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let isTimeComparisonActive: boolean;
  export let leaderboardMeasureNames: string[] = [];
  export let suppressTooltip: boolean;
  export let leaderboardShowContextForAllMeasures: boolean;
  export let leaderboardSortByMeasureName: string | null;
  export let isValidPercentOfTotal: (measureName: string) => boolean;
  export let dimensionColumnWidth: number;
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let formatters: Record<
    string,
    (value: number | string | null | undefined) => string | null | undefined
  >;

  function shouldShowContextColumns(measureName: string): boolean {
    return (
      leaderboardShowContextForAllMeasures ||
      measureName === leaderboardSortByMeasureName
    );
  }

  let hovered = false;
  let valueRect = new DOMRect(0, 0, DEFAULT_COLUMN_WIDTH);
  let deltaRect = new DOMRect(0, 0, COMPARISON_COLUMN_WIDTH);
  let parent: HTMLTableRowElement;

  $: ({
    dimensionValue,
    selectedIndex,
    values,
    pctOfTotals,
    prevValues,
    deltaRels,
    deltaAbs: deltaAbsMap,
    uri,
  } = itemData);

  $: selected = selectedIndex >= 0;

  const { getCellProps } = useCellInspector();

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: previousValueString =
    leaderboardMeasureNames.length === 1 &&
    prevValues[leaderboardMeasureNames[0]] !== undefined &&
    prevValues[leaderboardMeasureNames[0]] !== null
      ? formatters[leaderboardMeasureNames[0]]?.(
          prevValues[leaderboardMeasureNames[0]] as number,
        )
      : undefined;

  $: href = makeHref(uri, dimensionValue);

  $: deltaElementWidth = deltaRect?.width;
  $: valueElementWith = valueRect?.width;

  $: valueColumn.update(valueElementWith);
  $: deltaColumn.update(deltaElementWidth);

  $: barLengths = Object.fromEntries(
    Object.entries(pctOfTotals).map(([name, pct]) => [
      name,
      pct ? tableWidth * pct : 0,
    ]),
  );

  $: totalBarLength = Object.values(barLengths).reduce(
    (sum, length) => sum + length,
    0,
  );

  $: showZigZags = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      length > tableWidth,
    ]),
  );

  $: measureCellBarLengths = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      clamp(0, length, $valueColumn),
    ]),
  );

  $: barColor = excluded
    ? "rgb(243 244 246)"
    : selected || hovered
      ? "var(--color-primary-200)"
      : "var(--color-primary-100)";

  $: dimensionGradients =
    leaderboardMeasureNames.length === 1
      ? `linear-gradient(to right, ${barColor} ${Math.min(dimensionColumnWidth, totalBarLength)}px, transparent ${Math.min(dimensionColumnWidth, totalBarLength)}px)`
      : undefined;

  $: measureGradients =
    leaderboardMeasureNames.length === 1
      ? `linear-gradient(to right, ${barColor} ${Math.max(0, totalBarLength - dimensionColumnWidth)}px, transparent ${Math.max(0, totalBarLength - dimensionColumnWidth)}px)`
      : undefined;

  $: measureGradientMap =
    leaderboardMeasureNames.length === 1
      ? undefined
      : Object.fromEntries(
          leaderboardMeasureNames.map((name) => {
            const length = measureCellBarLengths[name];
            return [
              name,
              length
                ? `linear-gradient(to right, transparent 16px, ${barColor} 16px, ${barColor} ${length + 16}px, transparent ${length + 16}px)`
                : undefined,
            ];
          }),
        );

  $: showTooltip = hovered && !suppressTooltip;

  function shiftClickHandler(label: string) {
    let truncatedLabel = label?.toString();
    if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
      truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
    }
    copyToClipboard(
      label,
      `copied dimension value "${truncatedLabel}" to clipboard`,
    );
  }
</script>

<tr
  bind:this={parent}
  class:border-b={borderBottom}
  class:border-t={borderTop}
  class="relative"
  style:background={leaderboardMeasureNames.length === 1
    ? dimensionGradients
    : undefined}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  on:click={(e) => {
    if (e.shiftKey) return;
    toggleDimensionValueSelection(
      dimensionName,
      dimensionValue,
      false,
      e.ctrlKey || e.metaKey,
    );
  }}
>
  <td data-comparison-cell>
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={itemData?.selectedIndex}
    />
  </td>
  <td
    data-dimension-cell
    class:ui-copy={!atLeastOneActive}
    class:ui-copy-disabled={excluded}
    class:ui-copy-strong={!excluded && selected}
    on:click={modified({
      shift: () => shiftClickHandler(dimensionValue),
    })}
    class="relative size-full flex flex-none justify-between items-center leaderboard-label"
    style:background={dimensionGradients}
    {...getCellProps(dimensionValue)}
  >
    <span class="truncate">
      <FormattedDataType value={dimensionValue} truncate />
    </span>

    {#if previousValueString && hovered}
      <span
        class="opacity-50 whitespace-nowrap font-normal"
        transition:slide={{ axis: "x", duration: 200 }}
      >
        {previousValueString} →
      </span>
    {/if}

    {#if hovered && href}
      <a
        target="_blank"
        rel="noopener noreferrer"
        {href}
        title={href}
        on:click|stopPropagation
      >
        <ExternalLink className="fill-primary-600" />
      </a>
    {/if}
  </td>

  {#each Object.keys(values) as measureName}
    <td
      data-measure-cell
      on:click={modified({
        shift: () => shiftClickHandler(values[measureName]?.toString() || ""),
      })}
      style:background={leaderboardMeasureNames.length === 1
        ? measureGradients
        : measureGradientMap?.[measureName]}
      {...getCellProps(values[measureName]?.toString() || "")}
    >
      <div class="w-fit ml-auto bg-transparent" bind:contentRect={valueRect}>
        <FormattedDataType
          type="INTEGER"
          value={values[measureName]
            ? formatters[measureName]?.(values[measureName])
            : null}
        />
      </div>

      {#if showZigZags[measureName] && !isTimeComparisonActive && !isValidPercentOfTotal(measureName)}
        <LongBarZigZag />
      {/if}
    </td>

    {#if isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)}
      <td
        data-comparison-cell
        title={pctOfTotals[measureName]?.toString() || ""}
        on:click={modified({
          shift: () =>
            shiftClickHandler(pctOfTotals[measureName]?.toString() || ""),
        })}
      >
        <PercentageChange
          value={pctOfTotals[measureName]}
          color="text-gray-500"
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </td>
    {/if}

    {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
      <td
        data-comparison-cell
        title={deltaAbsMap[measureName]?.toString() || ""}
        on:click={modified({
          shift: () =>
            shiftClickHandler(deltaAbsMap[measureName]?.toString() || ""),
        })}
      >
        <FormattedDataType
          color="text-gray-500"
          type="INTEGER"
          value={deltaAbsMap[measureName]
            ? formatters[measureName]?.(deltaAbsMap[measureName])
            : null}
          customStyle={deltaAbsMap[measureName] !== null &&
          deltaAbsMap[measureName] < 0
            ? "text-red-500"
            : ""}
          truncate={true}
        />
      </td>
    {/if}

    {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
      <td
        data-comparison-cell
        title={deltaRels[measureName]?.toString() || ""}
        on:click={modified({
          shift: () =>
            shiftClickHandler(deltaRels[measureName]?.toString() || ""),
        })}
      >
        <PercentageChange
          value={deltaRels[measureName]
            ? formatMeasurePercentageDifference(deltaRels[measureName])
            : null}
          color="text-gray-500"
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </td>
    {/if}
  {/each}
</tr>

{#if showTooltip}
  {#await new Promise((r) => setTimeout(r, 600)) then}
    <FloatingElement
      target={parent}
      location="left"
      alignment="middle"
      distance={0}
      pad={0}
    >
      <LeaderboardTooltipContent
        {atLeastOneActive}
        {excluded}
        {filterExcludeMode}
        label={dimensionValue}
        {selected}
      />
    </FloatingElement>
  {/await}
{/if}

<style lang="postcss">
  td {
    @apply text-right p-0;
    @apply px-2 relative;
    height: 22px;
  }

  tr {
    @apply cursor-pointer;
    max-height: 22px;
  }

  tr:hover {
    @apply bg-gray-100;
  }

  td[data-comparison-cell] {
    @apply bg-surface px-1 truncate;
  }

  td[data-dimension-cell] {
    @apply sticky left-0 z-30 bg-surface;
  }

  tr:hover td[data-dimension-cell],
  tr:hover td[data-comparison-cell] {
    @apply bg-gray-100;
  }

  a {
    @apply absolute right-0 z-50 h-[22px] w-[32px];
    @apply bg-surface flex items-center justify-center shadow-md rounded-sm;
  }

  a:hover {
    @apply bg-primary-100;
  }

  td {
    height: 22px !important;
  }
</style>
