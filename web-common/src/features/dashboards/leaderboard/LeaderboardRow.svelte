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
  import type { LeaderboardItemData } from "./leaderboard-utils";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    DEFAULT_COL_WIDTH,
    deltaColumn,
    valueColumn,
  } from "./leaderboard-widths";
  import FloatingElement from "@rilldata/web-common/components/floating-element/FloatingElement.svelte";

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
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let formatter:
    | ((_value: number | undefined) => undefined)
    | ((value: string | number) => string);
  export let firstColumnWidth: number;
  export let suppressTooltip: boolean;

  let hovered = false;
  let valueRect = new DOMRect(0, 0, DEFAULT_COL_WIDTH);
  let deltaRect = new DOMRect(0, 0, DEFAULT_COL_WIDTH);
  let parent: HTMLTableRowElement;

  $: ({
    dimensionValue,
    selectedIndex,
    pctOfTotal,
    value,
    prevValue,
    deltaAbs,
    deltaRel,
    uri,
  } = itemData);

  $: selected = selectedIndex >= 0;

  $: formattedValue = value ? formatter(value) : null;
  $: formattedDeltaRel = deltaRel
    ? formatMeasurePercentageDifference(deltaRel)
    : null;
  $: formattedDelta = deltaAbs ? formatter(deltaAbs) : null;

  $: deltaElementWidth = deltaRect.width;
  $: valueElementWith = valueRect.width;

  $: valueColumn.update(valueElementWith);
  $: deltaColumn.update(deltaElementWidth);

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: previousValueString =
    prevValue !== undefined && prevValue !== null
      ? formatter(prevValue)
      : undefined;

  $: negativeChange = deltaAbs !== null && deltaAbs < 0;

  $: href = makeHref(uri);

  $: percentOfTotal = isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

  $: barLength = tableWidth * percentOfTotal;
  $: showZigZag = barLength > tableWidth;

  $: barColor = excluded
    ? "rgb(243 244 246)"
    : selected || hovered
      ? "var(--color-primary-200)"
      : "var(--color-primary-100)";

  $: secondCellBarLength = clamp(0, barLength - firstColumnWidth, $valueColumn);
  $: thirdCellBarLength = isTimeComparisonActive
    ? clamp(0, barLength - firstColumnWidth - $valueColumn, $deltaColumn)
    : isValidPercentOfTotal
      ? clamp(0, barLength - firstColumnWidth - $valueColumn, DEFAULT_COL_WIDTH)
      : 0;
  $: fourthCellBarLength = isTimeComparisonActive
    ? clamp(
        0,
        barLength - firstColumnWidth - $valueColumn - $deltaColumn,
        DEFAULT_COL_WIDTH,
      )
    : 0;

  // Update the gradients
  $: firstCellGradient = `linear-gradient(to right, ${barColor}
    ${barLength}px, transparent ${barLength}px)`;

  $: secondCellGradient = secondCellBarLength
    ? `linear-gradient(to right, ${barColor}
    ${secondCellBarLength}px, transparent ${secondCellBarLength}px)`
    : undefined;

  $: thirdCellGradient = thirdCellBarLength
    ? `linear-gradient(to right, ${barColor}
    ${thirdCellBarLength}px, transparent ${thirdCellBarLength}px)`
    : undefined;

  $: fourthCellGradient = fourthCellBarLength
    ? `linear-gradient(to right, ${barColor}
    ${fourthCellBarLength}px, transparent ${fourthCellBarLength}px)`
    : undefined;

  // uri template or "true" string literal or undefined
  function makeHref(uriTemplateOrBoolean: string | boolean | null) {
    if (!uriTemplateOrBoolean) {
      return undefined;
    }

    // temporary fix where uriTemplateOrBoolean is coming in as 0/1 instead of false/true
    if (typeof uriTemplateOrBoolean === "number") {
      uriTemplateOrBoolean = Boolean(uriTemplateOrBoolean);
    }

    const uri =
      uriTemplateOrBoolean === true
        ? dimensionValue
        : uriTemplateOrBoolean.replace(/\s/g, "");

    const hasProtocol = /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(uri);

    if (!hasProtocol) {
      return "https://" + uri;
    } else {
      return uri;
    }
  }

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

  <td
    style:background={secondCellGradient}
    on:click={modified({
      shift: () => shiftClickHandler(value?.toString() || ""),
    })}
  >
    <div class="w-fit ml-auto bg-transparent" bind:contentRect={valueRect}>
      <FormattedDataType type="INTEGER" value={formattedValue} />
    </div>

    {#if showZigZag && !isTimeComparisonActive && !isValidPercentOfTotal}
      <LongBarZigZag />
    {/if}
  </td>

  {#if isTimeComparisonActive || isValidPercentOfTotal}
    <td
      style:background={thirdCellGradient}
      on:click={modified({
        shift: () => shiftClickHandler(deltaAbs?.toString() || ""),
      })}
    >
      {#if isTimeComparisonActive}
        <div class="w-fit ml-auto" bind:contentRect={deltaRect}>
          <FormattedDataType
            type="INTEGER"
            value={formattedDelta}
            customStyle={negativeChange ? "text-red-500" : ""}
          />
        </div>
      {:else}
        <PercentageChange value={pctOfTotal} />
        {#if showZigZag}
          <LongBarZigZag />
        {/if}
      {/if}
    </td>
  {/if}

  {#if isTimeComparisonActive}
    <td
      style:background={fourthCellGradient}
      on:click={modified({
        shift: () => shiftClickHandler(deltaRel?.toString() || ""),
      })}
    >
      <PercentageChange value={formattedDeltaRel} />
      {#if showZigZag}
        <LongBarZigZag />
      {/if}
    </td>
  {/if}
</tr>

{#if hovered && !suppressTooltip}
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

  a {
    @apply absolute right-0 z-50 h-[22px] w-[32px];
    @apply bg-surface flex items-center justify-center shadow-md rounded-sm;
  }

  a:hover {
    @apply bg-primary-100;
  }
</style>
