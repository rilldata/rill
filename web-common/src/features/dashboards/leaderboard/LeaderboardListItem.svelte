<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fly, slide } from "svelte/transition";
  import BarAndLabel from "../../../components/BarAndLabel.svelte";
  import { FormattedDataType } from "@rilldata/web-common/components/data-types";

  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";

  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  import LeaderboardEntryTooltip from "./LeaderboardEntryTooltip.svelte";

  import PercentageChange from "../../../components/data-types/PercentageChange.svelte";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";

  export let measureValue: number;
  // export let color = "bg-blue-200 dark:bg-blue-600";
  export let isActive = false;
  export let excluded = false;
  export let showContext = false;

  export let loading = false;
  export let atLeastOneActive = false;
  export let label: string | number;
  export let previousValueString: string;
  export let formattedValue: string;
  export let percentChangeFormatted;
  export let filterExcludeMode;

  /** if this value is a summable measure, we'll show the bar. Otherwise, don't. */
  export let isSummableMeasure;
  /** for summable measures, this is the value we use to calculate the bar % to fill */
  export let referenceValue;

  /** compact mode is used in e.g. profiles */
  export let compact = false;

  const dispatch = createEventDispatcher();

  let hovered = false;
  const onHover = () => {
    hovered = true;
    dispatch("focus");
  };
  const onLeave = () => {
    hovered = false;
    dispatch("blur");
  };
  /** used for overly-large bar values */
  let zigZag =
    "M" +
    Array.from({ length: 7 })
      .map((_, i) => {
        return `${15 - 4 * (i % 2)} ${1.7 * (i * 2)}`;
      })
      .join(" L");

  $: height = compact ? "18px" : "22px";

  let renderedBarValue = 0; // should be between 0 and 1.
  $: {
    renderedBarValue = isSummableMeasure
      ? referenceValue
        ? measureValue / referenceValue
        : 0
      : 0;
    // if this somehow creates an NaN, let's set it to 0.
    renderedBarValue = !isNaN(renderedBarValue) ? renderedBarValue : 0;
  }
  $: color = excluded
    ? "ui-measure-bar-excluded"
    : isActive
    ? "ui-measure-bar-included-selected"
    : "ui-measure-bar-included";

  $: console.log("color", color, "renderedBarValue", renderedBarValue);
</script>

<Tooltip location="right">
  <button
    class="block flex flex-row w-full text-left transition-color"
    on:blur={onLeave}
    on:click
    on:focus={onHover}
    on:mouseleave={onLeave}
    on:mouseover={onHover}
    transition:slide|local={{ duration: 200 }}
  >
    <LeaderboardItemFilterIcon {isActive} {excluded} />
    <BarAndLabel
      {color}
      justify={false}
      showBackground={false}
      showHover
      tweenParameters={{ duration: 200 }}
      value={renderedBarValue}
    >
      <div class="grid leaderboard-entry items-center gap-x-3" style:height>
        <div
          class="justify-self-start text-left w-full text-ellipsis overflow-hidden whitespace-nowrap"
        >
          <div
            class:ui-copy={!atLeastOneActive && !loading}
            class:ui-copy-strong={!excluded && isActive}
            class:ui-copy-disabled={excluded}
            class="w-full text-ellipsis overflow-hidden whitespace-nowrap"
          >
            <FormattedDataType value={label} />
          </div>
        </div>
        <div
          class="justify-self-end overflow-hidden ui-copy-number flex gap-x-4 items-baseline"
        >
          <div
            class="flex items-baseline gap-x-1"
            class:ui-copy-disabled={excluded}
            class:ui-copy-strong={!excluded && isActive}
            in:fly={{ duration: 200, y: 4 }}
          >
            {#if previousValueString}
              <span
                class="inline-block opacity-50"
                transition:slideRight={{ duration: LIST_SLIDE_DURATION }}
              >
                {previousValueString}
                →
              </span>
            {/if}
            <FormattedDataType
              type="INTEGER"
              value={formattedValue || measureValue}
            />
          </div>
          {#if showContext}
            <div
              class:ui-copy-disabled={excluded}
              class:ui-copy-strong={!excluded && isActive}
              class="text-xs text-gray-500 dark:text-gray-400"
              style:width="44px"
            >
              <PercentageChange value={percentChangeFormatted} />
            </div>
          {/if}
        </div>
      </div>
    </BarAndLabel>
  </button>
  <!-- if the value is greater than 100%, we should add this little serration -->
  {#if renderedBarValue > 1.001}
    <div
      style="position: relative"
      transition:fly|local={{ duration: 200, x: 20 }}
    >
      <svg
        style="
    position:absolute;
    right: 0px;
    transform: translateY(-{height});
  "
        width="15"
        height="22"
      >
        <path d={zigZag} fill="white" />
      </svg>
    </div>
  {/if}

  <TooltipContent slot="tooltip-content">
    <LeaderboardEntryTooltip
      {label}
      {atLeastOneActive}
      {excluded}
      {filterExcludeMode}
    />
  </TooltipContent>
</Tooltip>

<style>
  .leaderboard-entry {
    grid-template-columns: auto max-content;
  }
</style>
