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
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    DEFAULT_COL_WIDTH,
    deltaColumn,
    valueColumn,
  } from "./leaderboard-widths";
  import FloatingElement from "@rilldata/web-common/components/floating-element/FloatingElement.svelte";
  import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
  export let tableWidth: number;
  export let borderTop = false;
  export let borderBottom = false;
  export let isSummableMeasure: boolean;
  export let isBeingCompared: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let isValidPercentOfTotal: boolean;
  export let isTimeComparisonActive: boolean;
  export let contextColumns: string[] = [];
  export let activeMeasureNames: string[] = [];
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let formatter:
    | ((_value: number | undefined) => undefined)
    | ((value: string | number) => string);
  export let dimensionColumnWidth: number;
  export let suppressTooltip: boolean;

  let hovered = false;
  let valueRect = new DOMRect(0, 0, DEFAULT_COL_WIDTH);
  let deltaRect = new DOMRect(0, 0, DEFAULT_COL_WIDTH);
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

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: previousValueString =
    activeMeasureNames.length === 1 &&
    prevValues[activeMeasureNames[0]] !== undefined &&
    prevValues[activeMeasureNames[0]] !== null
      ? formatter(prevValues[activeMeasureNames[0]] as number)
      : undefined;

  $: href = makeHref(uri, dimensionValue);

  $: deltaElementWidth = deltaRect.width;
  $: valueElementWith = valueRect.width;

  $: valueColumn.update(valueElementWith);
  $: deltaColumn.update(deltaElementWidth);

  $: barColor = excluded
    ? "rgb(243 244 246)"
    : selected || hovered
      ? "var(--color-primary-200)"
      : "var(--color-primary-100)";

  $: barLengths = Object.fromEntries(
    Object.entries(pctOfTotals).map(([name, pct]) => [
      name,
      isSummableMeasure && pct ? tableWidth * pct : 0,
    ]),
  );

  $: showZigZags = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      length > tableWidth,
    ]),
  );

  $: secondCellBarLengths = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      clamp(0, length - dimensionColumnWidth, $valueColumn),
    ]),
  );

  $: thirdCellBarLengths = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      isTimeComparisonActive
        ? clamp(0, length - dimensionColumnWidth - $valueColumn, $deltaColumn)
        : isValidPercentOfTotal
          ? clamp(
              0,
              length - dimensionColumnWidth - $valueColumn,
              DEFAULT_COL_WIDTH,
            )
          : 0,
    ]),
  );

  $: fourthCellBarLengths = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      clamp(
        0,
        length -
          dimensionColumnWidth -
          $valueColumn -
          (thirdCellBarLengths[name] || 0),
        DEFAULT_COL_WIDTH,
      ),
    ]),
  );

  $: fifthCellBarLengths = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      clamp(
        0,
        length -
          dimensionColumnWidth -
          $valueColumn -
          (thirdCellBarLengths[name] || 0) -
          (fourthCellBarLengths[name] || 0),
        DEFAULT_COL_WIDTH,
      ),
    ]),
  );

  $: firstCellGradient =
    activeMeasureNames.length >= 2
      ? "bg-white"
      : `linear-gradient(to right, ${barColor}
    ${Math.max(...Object.values(barLengths))}px, transparent ${Math.max(...Object.values(barLengths))}px)`;

  $: secondCellGradients =
    activeMeasureNames.length === 1
      ? "bg-white"
      : Object.fromEntries(
          Object.entries(secondCellBarLengths).map(([name, length]) => [
            name,
            length
              ? `linear-gradient(to right, transparent 16px, ${barColor} 16px,
    ${barColor} ${length + 16}px, transparent ${length + 16}px)`
              : undefined,
          ]),
        );

  $: thirdCellGradients = Object.fromEntries(
    Object.entries(thirdCellBarLengths).map(([name, length]) => [
      name,
      length
        ? `linear-gradient(to right, ${barColor}
    ${length}px, transparent ${length}px)`
        : undefined,
    ]),
  );

  $: fourthCellGradients = Object.fromEntries(
    Object.entries(fourthCellBarLengths).map(([name, length]) => [
      name,
      length
        ? `linear-gradient(to right, ${barColor}
    ${length}px, transparent ${length}px)`
        : undefined,
    ]),
  );

  $: fifthCellGradients = Object.fromEntries(
    Object.entries(fifthCellBarLengths).map(([name, length]) => [
      name,
      length
        ? `linear-gradient(to right, ${barColor}
    ${length}px, transparent ${length}px)`
        : undefined,
    ]),
  );

  $: showDeltaAbsolute =
    isTimeComparisonActive &&
    contextColumns.includes(LeaderboardContextColumn.DELTA_ABSOLUTE);

  $: showDeltaPercent =
    isTimeComparisonActive &&
    contextColumns.includes(LeaderboardContextColumn.DELTA_PERCENT);

  $: showPercentOfTotal =
    isTimeComparisonActive &&
    isValidPercentOfTotal &&
    contextColumns.includes(LeaderboardContextColumn.PERCENT);

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
  <td>
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={itemData?.selectedIndex}
    />
  </td>
  <td
    data-first-cell
    style:background={firstCellGradient}
    class:ui-copy={!atLeastOneActive}
    class:ui-copy-disabled={excluded}
    class:ui-copy-strong={!excluded && selected}
    on:click={modified({
      shift: () => shiftClickHandler(dimensionValue),
    })}
    class="relative size-full flex flex-none justify-between items-center leaderboard-label"
  >
    <FormattedDataType value={dimensionValue} truncate />

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
      data-second-cell
      style:background={secondCellGradients[measureName]}
      on:click={modified({
        shift: () => shiftClickHandler(values[measureName]?.toString() || ""),
      })}
    >
      <div class="w-fit ml-auto bg-transparent" bind:contentRect={valueRect}>
        <FormattedDataType
          type="INTEGER"
          value={values[measureName] ? formatter(values[measureName]) : null}
        />
      </div>

      {#if showZigZags[measureName] && !isTimeComparisonActive && !isValidPercentOfTotal}
        <LongBarZigZag />
      {/if}
    </td>

    {#if showDeltaAbsolute}
      <td
        data-third-cell
        style:background={thirdCellGradients[measureName]}
        on:click={modified({
          shift: () =>
            shiftClickHandler(deltaAbsMap[measureName]?.toString() || ""),
        })}
      >
        {#if isTimeComparisonActive}
          <div class="w-fit ml-auto" bind:contentRect={deltaRect}>
            <FormattedDataType
              type="INTEGER"
              value={deltaAbsMap[measureName]
                ? formatter(deltaAbsMap[measureName])
                : null}
              customStyle={deltaAbsMap[measureName] !== null &&
              deltaAbsMap[measureName] < 0
                ? "text-red-500"
                : ""}
            />
          </div>
        {:else}
          <PercentageChange value={pctOfTotals[measureName]} />
          {#if showZigZags[measureName]}
            <LongBarZigZag />
          {/if}
        {/if}
      </td>
    {/if}

    {#if showDeltaPercent}
      <td
        data-fourth-cell
        style:background={fourthCellGradients[measureName]}
        on:click={modified({
          shift: () =>
            shiftClickHandler(deltaRels[measureName]?.toString() || ""),
        })}
      >
        <PercentageChange
          value={deltaRels[measureName]
            ? formatMeasurePercentageDifference(deltaRels[measureName])
            : null}
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </td>
    {/if}

    {#if showPercentOfTotal}
      <td
        data-fifth-cell
        style:background={fifthCellGradients[measureName]}
        on:click={modified({
          shift: () =>
            shiftClickHandler(pctOfTotals[measureName]?.toString() || ""),
        })}
      >
        <PercentageChange value={pctOfTotals[measureName]} />
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
  }

  tr:hover {
    @apply bg-gray-100;
  }

  td:first-of-type {
    @apply p-0 bg-surface;
  }

  /* 
  td:nth-of-type(3) {
    @apply sticky left-0 z-20;
    
  } */

  /* tr:hover td:nth-of-type(2) {
    @apply bg-gray-100;
  } */

  /* 
  td:nth-of-type(2)::after {
    content: "";
    @apply absolute right-0 top-0 bottom-0 w-px bg-gray-200;
  } */

  a {
    @apply absolute right-0 z-50 h-[22px] w-[32px];
    @apply bg-surface flex items-center justify-center shadow-md rounded-sm;
  }

  a:hover {
    @apply bg-primary-100;
  }
</style>
