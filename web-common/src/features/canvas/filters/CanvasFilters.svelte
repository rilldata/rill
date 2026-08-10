<script lang="ts">
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { getPanRangeForTimeRange } from "@rilldata/web-common/features/dashboards/state-managers/selectors/charts";
  import SuperPill from "@rilldata/web-common/features/dashboards/time-controls/super-pill/SuperPill.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import Metadata from "../../dashboards/time-controls/super-pill/components/Metadata.svelte";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import ExpressionFilters from "../../dashboards/filters/ExpressionFilters.svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";

  let {
    maxWidth,
    canvasName,
  }: {
    readOnly?: boolean;
    maxWidth: number;
    canvasName: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let showDefaultItem = false;

  let {
    canvasEntity: {
      timeManager: {
        state: {
          comparisonIntervalStore,
          showTimeComparisonStore,
          timeZoneStore,
          canPanStore,
          set,
          rangeStore,
          grainStore,
          comparisonRangeStore,
          interval: intervalStore,
          minMaxTimeStamps,
        },
        hasTimeSeriesStore,
        largestMinTimeGrain,
        defaultTimeRangeStore,
        timeRangeOptionsStore,
        availableTimeZonesStore,
        allowCustomRangeStore,
        maxQueryTimeRangeStore,
      },
      expressionFilterManager,
    },
  } = $derived(getCanvasStore(canvasName, runtimeClient.instanceId));
  // svelte-ignore state_referenced_locally
  expressionFilterManager.createListener();

  function updateUrlParams(params: URLSearchParams) {
    expressionFilterManager.setUrlParams(params);
  }
  $effect(() => updateUrlParams(page.url.searchParams));
  // svelte-ignore state_referenced_locally
  expressionFilterManager.on("state-changed", () => {
    const newUrl = new URL(page.url);
    expressionFilterManager.applyFilterToParams(newUrl.searchParams);
    if (newUrl.search === page.url.search) return;
    void goto(newUrl);
  });

  let selectedRange = $derived($rangeStore);
  let interval = $derived($intervalStore);
  let minTimeGrain = $derived($largestMinTimeGrain);

  let timeStart = $derived(interval?.start.toISO());
  let timeEnd = $derived(interval?.end.toISO());

  let minMax = $derived($minMaxTimeStamps);

  let hasTimeSeries = $derived($hasTimeSeriesStore);

  let minDate = $derived(minMax?.min);
  let maxDate = $derived(minMax?.max);
  let showTimeComparison = $derived($showTimeComparisonStore);

  let activeTimeZone = $derived($timeZoneStore);
  let comparisonInterval = $derived($comparisonIntervalStore);
  let comparisonRange = $derived($comparisonRangeStore);

  let activeTimeGrain = $derived($grainStore);
  let defaultTimeRange = $derived($defaultTimeRangeStore);
  let availableTimeZones = $derived($availableTimeZonesStore);
  let timeRanges = $derived($timeRangeOptionsStore);
  let allowCustomTimeRange = $derived($allowCustomRangeStore);
  let maxQueryTimeRange = $derived($maxQueryTimeRangeStore);

  let canPan = $derived($canPanStore);

  function onPan(direction: "left" | "right") {
    if (!interval || !selectedRange) return;
    const getPanRange = getPanRangeForTimeRange(
      {
        end: interval?.end.toJSDate(),
        start: interval?.start.toJSDate(),
        name: selectedRange,
      },
      activeTimeZone,
    );
    const panRange = getPanRange(direction);

    if (!panRange) return;
    const { start, end } = panRange;

    if (!activeTimeGrain) return;

    set.range(`${start.toISOString()},${end.toISOString()}`);
  }
</script>

<div
  role="presentation"
  class="flex flex-col gap-y-2 size-full pointer-events-none"
  style:max-width="{maxWidth}px"
>
  {#if hasTimeSeries}
    <div class="p-2 flex justify-between size-full py-0">
      <div class="flex items-center size-full">
        <div class="flex-none h-full pt-1.5 pointer-events-auto">
          <Tooltip.Root delayDuration={0}>
            <Tooltip.Trigger class="cursor-default">
              <Calendar size="16px" />
            </Tooltip.Trigger>
            <Tooltip.Content side="bottom" sideOffset={10} class="z-50">
              <Metadata
                timeZone={activeTimeZone}
                timeStart={minDate?.toJSDate()}
                timeEnd={maxDate?.toJSDate()}
              />
            </Tooltip.Content>
          </Tooltip.Root>
        </div>
        <div
          class="flex flex-wrap gap-x-2 gap-y-1.5 pl-2 pointer-events-auto size-full pr-2"
        >
          <SuperPill
            context={canvasName}
            {minDate}
            {maxDate}
            selectedRangeAlias={selectedRange}
            showPivot={false}
            {minTimeGrain}
            {defaultTimeRange}
            {availableTimeZones}
            {timeRanges}
            complete={false}
            {interval}
            {timeStart}
            {timeEnd}
            {activeTimeGrain}
            {activeTimeZone}
            canPanLeft={canPan.left}
            canPanRight={canPan.right}
            watermark={undefined}
            {allowCustomTimeRange}
            {maxQueryTimeRange}
            {showDefaultItem}
            applyRange={(timeRange) => {
              const string = `${timeRange.start.toISOString()},${timeRange.end.toISOString()}`;
              set.range(string);
            }}
            onSelectRange={set.range}
            onTimeGrainSelect={set.grain}
            onSelectTimeZone={set.zone}
            {onPan}
          />
          <CanvasComparisonPill
            {minDate}
            {maxDate}
            {comparisonInterval}
            {comparisonRange}
            {interval}
            {selectedRange}
            {activeTimeGrain}
            {activeTimeZone}
            {minTimeGrain}
            {showTimeComparison}
            {allowCustomTimeRange}
            onDisplayTimeComparison={set.comparison}
            onSetSelectedComparisonRange={(range) => {
              if (range.name === "CUSTOM_COMPARISON_RANGE") {
                const stringRange = `${range.start.toISOString()},${range.end.toISOString()}`;
                set.comparison(stringRange);
              } else if (range.name) {
                set.comparison(range.name);
              }
            }}
          />
        </div>
      </div>
    </div>
  {/if}

  <div class="pointer-events-auto ml-2">
    <ExpressionFilters
      {expressionFilterManager}
      {timeStart}
      {timeEnd}
      timeControlsReady={!hasTimeSeries || !!interval}
    />
  </div>
</div>
