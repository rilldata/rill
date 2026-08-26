<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import { Nudge } from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components";
  import TimeRangePicker from "@rilldata/web-common/features/dashboards/time-controls/TimeRangePicker.svelte";
  import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Metadata from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/Metadata.svelte";
  import ComparisonTimeRangePicker from "@rilldata/web-common/features/dashboards/time-controls/ComparisonTimeRangePicker.svelte";
  import { type V1ExploreTimeRange } from "@rilldata/web-common/runtime-client";
  import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

  let {
    timeFilterManager,
    metricsViewsProvider,
    yamlConfigProvider,
    context,
    config,
  }: {
    timeFilterManager: TimeFilterManager;
    metricsViewsProvider: MetricsViewsProvider;
    yamlConfigProvider: YAMLConfigProvider;
    dimensions: string[];
    timeRanges: V1ExploreTimeRange[];
    context: string;
    config: TimeFiltersConfig;
  } = $props();
  let hidePan = $derived(config.hidePan);
  let canPanLeft = $derived(config.canPanLeft ?? !hidePan);
  let canPanRight = $derived(config.canPanRight ?? !hidePan);

  let timeRangeManager = $derived(timeFilterManager.timeRangeManager);
  let { timeZone, minDate, maxDate } = $derived(timeRangeManager);

  let comparisonTimeRangeManager = $derived(
    timeFilterManager.comparisonTimeRangeManager,
  );

  function onPan() {
    // TODO
  }
</script>

<div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
  <Tooltip.Root delayDuration={0}>
    <Tooltip.Trigger class="cursor-default text-fg-secondary">
      <Calendar size="16px" />
    </Tooltip.Trigger>
    <Tooltip.Content side="bottom" sideOffset={10}>
      <Metadata
        {timeZone}
        timeStart={minDate?.toJSDate()}
        timeEnd={maxDate?.toJSDate()}
      />
    </Tooltip.Content>
  </Tooltip.Root>

  <div class="wrapper">
    {#if !hidePan}
      <Nudge {canPanLeft} {canPanRight} {onPan} direction="left" />
      <Nudge {canPanLeft} {canPanRight} {onPan} direction="right" />
    {/if}

    <TimeRangePicker
      {timeRangeManager}
      {metricsViewsProvider}
      {yamlConfigProvider}
      {context}
      {config}
    />

    <ComparisonTimeRangePicker
      {timeRangeManager}
      {comparisonTimeRangeManager}
      {metricsViewsProvider}
      {config}
    />
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply flex w-fit;
    @apply h-[26px] rounded-full;
    @apply overflow-hidden;
  }

  /* Match both direct child buttons and buttons inside display:contents wrappers */
  :global(.wrapper > button),
  :global(.wrapper > :not(button) > button) {
    @apply border text-fg-primary -ml-px cursor-pointer;
    @apply px-2 flex items-center justify-center bg-surface-background;
  }

  :global(.wrapper > button:first-child),
  :global(.wrapper > :first-child > button:first-child) {
    @apply ml-0 pl-2.5 rounded-l-full;
  }

  :global(.wrapper > button:last-child),
  :global(.wrapper > :last-child > button:last-child) {
    @apply pr-2.5 rounded-r-full;
  }

  :global(.wrapper > button:focus),
  :global(.wrapper > :not(button) > button:focus) {
    @apply z-50;
  }

  :global(.wrapper > button:hover:not(:disabled)),
  :global(.wrapper > :not(button) > button:hover:not(:disabled)) {
    @apply bg-surface-hover;
  }
</style>
