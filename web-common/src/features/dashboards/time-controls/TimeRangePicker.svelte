<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import type { TimeRangeManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeRangeManager.svelte.ts";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import PrimaryRangeTooltip from "@rilldata/web-common/features/dashboards/time-controls/super-pill/new-time-dropdown/PrimaryRangeTooltip.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    ALL_TIME_RANGE_ALIAS,
    bucketYamlRanges,
    getRangeLabel,
    RILL_TO_LABEL,
  } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
  import { getAbbreviationForIANA } from "@rilldata/web-common/lib/time/timezone";
  import { DateTime } from "luxon";
  import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import TimeRangeSearch from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/TimeRangeSearch.svelte";
  import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
  import TimeRangeOptionGroup from "@rilldata/web-common/features/dashboards/time-controls/super-pill/new-time-dropdown/TimeRangeOptionGroup.svelte";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import SyntaxElement from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/SyntaxElement.svelte";
  import ZoneContent from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/ZoneContent.svelte";
  import { Clock, Check } from "lucide-svelte";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains.ts";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import CalendarPlusDateInput from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/CalendarPlusDateInput.svelte";
  import {
    RillAllTimeInterval,
    RillIsoInterval,
    RillPeriodToGrainInterval,
  } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
  import TruncationSelector from "@rilldata/web-common/features/dashboards/time-controls/super-pill/new-time-dropdown/TruncationSelector.svelte";
  import type { TimeFiltersConfig } from "@rilldata/web-common/features/dashboards/time-controls/time-filters-config.ts";
  import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
  import { getTimeDimensionOptions } from "@rilldata/web-common/features/dashboards/time-controls/time-range-utils.ts";

  let {
    timeRangeManager,
    metricsViewsProvider,
    yamlConfigProvider,
    context,
    config,
  }: {
    timeRangeManager: TimeRangeManager;
    metricsViewsProvider: MetricsViewsProvider;
    yamlConfigProvider: YAMLConfigProvider;
    context: string;
    config: TimeFiltersConfig;
  } = $props();
  let {
    showTimeDimensionSelector = false,
    allowCustomTimeRange = true,
    showDefaultItem,
    lockTimeZone = false,
    showFullRange = true,
    showWatermark = false,
  } = $derived(config);

  let {
    timeRange: timeString,
    timeGrain,
    timeZone,
    timeDimension,

    minDate,
    maxDate,

    interval,
    parsedTime,
    truncationGrain,
    ref,
    snapToEnd,

    onSelectRange,
    onSelectZone,
    onSelectGrain,
    onSelectAsOfOption,
    onSelectTimeDimension,
  } = $derived(timeRangeManager);

  let { smallestTimeGrain, maxQueryTimeRange } = $derived(metricsViewsProvider);

  let {
    restrictedDimensions,
    primaryTimeDimension,
    defaultTimeRange,
    timeRanges,
    timeZones,
  } = $derived(yamlConfigProvider);

  let rangeBuckets = $derived(
    bucketYamlRanges(timeRanges, smallestTimeGrain, true, maxQueryTimeRange),
  );

  let timeDimensions = $derived(
    getTimeDimensionOptions(
      metricsViewsProvider.dimensions,
      restrictedDimensions,
    ),
  );

  let watermark = $derived(
    showWatermark && metricsViewsProvider.timeRangeSummary?.watermark
      ? DateTime.fromISO(metricsViewsProvider.timeRangeSummary.watermark)
      : undefined,
  );

  let open = $state(false);
  // svelte-ignore state_referenced_locally
  let searchValue: string | undefined = $state(timeString);
  let timeZonePickerOpen = $state(false);
  let timeAxisPickerOpen = $state(false);
  let showCalendarPicker = $state(false);

  let allTimeAllowed = $derived(
    !(maxQueryTimeRange && maxQueryTimeRange.as("milliseconds") > 0),
  );
  let selectedLabel = $derived(getRangeLabel(timeString));
  let zoneAbbreviation = $derived(
    getAbbreviationForIANA(maxDate ?? DateTime.now(), timeZone),
  );
  let dateTimeAnchor = $derived(returnAnchor(ref, timeZone));

  let usingLegacyTime = $derived(parsedTime?.isOldFormat);
  let hideTruncationSelector = $derived(
    parsedTime?.interval instanceof RillIsoInterval ||
      parsedTime?.interval instanceof RillAllTimeInterval,
  );

  let timeColumn = $derived(timeDimension || primaryTimeDimension);
  let activeTimeDimension = $derived(
    timeDimensions.find((d) => d.value === timeColumn),
  );

  // Zone is taken as a param to make it reactive
  function returnAnchor(
    asOf: string | undefined,
    zone: string,
  ): DateTime | undefined {
    if (!maxDate) return DateTime.now().setZone(zone);
    if (asOf === "latest") {
      return maxDate.setZone(zone);
    } else if (asOf === "watermark" && watermark) {
      return watermark.setZone(zone);
    } else if (asOf === "now" || !asOf) {
      return DateTime.now().setZone(zone);
    }
  }
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.metaKey && e.key === "k") {
      open = !open;
    }
  }}
/>

<Popover.Root
  bind:open
  onOpenChange={(o) => {
    if (o) {
      searchValue = timeString;
    }
  }}
>
  <Tooltip.Root delayDuration={800} ignoreNonKeyboardFocus>
    <Tooltip.Trigger>
      {#snippet child({ props: tooltipProps })}
        <Popover.Trigger>
          {#snippet child({ props: popoverProps })}
            <button
              {...tooltipProps}
              {...popoverProps}
              class="flex gap-x-1.5"
              aria-label={m.dashboard_select_time_range()}
              type="button"
            >
              {#if timeString}
                <b class="line-clamp-1 flex-none">
                  {#if selectedLabel?.startsWith("-") || !isNaN(Number(selectedLabel?.[0]))}
                    {m.dashboard_custom()}
                  {:else}
                    {selectedLabel}
                  {/if}
                </b>
              {/if}

              {#if showFullRange}
                <RangeDisplay {interval} {timeGrain} />

                <div
                  class="font-bold bg-surface-muted rounded-[2px] p-1 py-0 text-fg-secondary text-[11px]"
                >
                  {zoneAbbreviation}
                </div>
              {/if}

              <span
                class="flex-none transition-transform"
                class:-rotate-180={open}
              >
                <CaretDownIcon />
              </span>
            </button>
          {/snippet}
        </Popover.Trigger>
      {/snippet}
    </Tooltip.Trigger>

    <Tooltip.Content side="bottom" sideOffset={8} class="z-50">
      {#if interval}
        <PrimaryRangeTooltip
          {timeString}
          {interval}
          timeDimensionLabel={activeTimeDimension?.label}
          timeDimensionDescription={activeTimeDimension?.description}
        />
      {/if}
    </Tooltip.Content>
  </Tooltip.Root>

  <Popover.Content
    align="start"
    class="p-0 w-fit overflow-hidden flex flex-col"
  >
    <TimeRangeSearch
      inError={!parsedTime && !!timeString && !usingLegacyTime}
      width={showCalendarPicker ? 456 : 224}
      {context}
      {timeString}
      bind:searchValue
      onSelectRange={(range) => {
        open = false;
        void onSelectRange(range);
      }}
    />

    <div
      class="flex w-56 max-h-fit"
      class:!w-[456px]={showCalendarPicker}
      style:height="470px"
    >
      <div
        class="flex flex-col w-56 overflow-y-auto overflow-x-hidden flex-none py-1"
      >
        <div class="overflow-x-hidden">
          {#if showDefaultItem && defaultTimeRange}
            <TimeRangeOptionGroup
              {timeString}
              options={[parseRillTime(defaultTimeRange)]}
              onClick={onSelectRange}
            />
          {/if}

          <TimeRangeOptionGroup
            {timeString}
            options={rangeBuckets.custom}
            onClick={onSelectRange}
          />

          <TimeRangeOptionGroup
            {timeString}
            options={rangeBuckets.latest}
            onClick={onSelectRange}
          />

          <TimeRangeOptionGroup
            {timeString}
            options={rangeBuckets.periodToDate}
            onClick={onSelectRange}
          />

          <TimeRangeOptionGroup
            {timeString}
            options={rangeBuckets.previous}
            onClick={(r) => void onSelectRange(r, true)}
          />

          {#if allTimeAllowed}
            <div class="w-full h-fit px-1">
              <button
                type="button"
                role="menuitem"
                class="group truncate h-7 p-2 text-popover-foreground justify-between overflow-hidden hover:bg-popover-accent rounded-sm w-full select-none flex items-center"
                onclick={() => void onSelectRange("inf")}
              >
                <span class:font-bold={timeString === ALL_TIME_RANGE_ALIAS}>
                  {RILL_TO_LABEL[ALL_TIME_RANGE_ALIAS]}
                </span>
              </button>
            </div>
          {/if}
        </div>

        {#if allowCustomTimeRange}
          <div class="w-full h-fit px-1">
            <div class="h-px w-full bg-border my-1"></div>
            <button
              type="button"
              role="menuitem"
              class:font-bold={false}
              onclick={() => {
                showCalendarPicker = !showCalendarPicker;
              }}
              class="truncate text-fg-primary w-full text-left gap-x-1 pr-1 hover:bg-popover-accent flex items-center flex-shrink pl-2 h-7 rounded-sm"
            >
              <Calendar size="14px" />
              <div class="mr-auto">{m.dashboard_custom()}</div>

              <CaretDownIcon
                className="-rotate-90 text-fg-secondary"
                size="14px"
              />
            </button>
          </div>
        {/if}

        {#if !lockTimeZone}
          <div class="w-full h-fit px-1">
            <div class="h-px w-full bg-border my-1"></div>

            <Popover.Root bind:open={timeZonePickerOpen}>
              <Popover.Trigger
                onclick={() => {
                  showCalendarPicker = false;
                }}
                class="group h-7 overflow-hidden hover:bg-popover-accent flex-none rounded-sm w-full select-none flex items-center truncate text-left gap-x-1 pr-1 pl-2"
              >
                <div class="flex-none">
                  <Globe size="14px" className="text-fg-primary" />
                </div>
                <div class="mr-auto text-fg-primary">
                  {m.dashboard_time_zone()}
                </div>
                <div class="sr-only group-hover:not-sr-only">
                  <SyntaxElement range={zoneAbbreviation} />
                </div>
                <CaretDownIcon
                  className="-rotate-90 text-fg-secondary"
                  size="14px"
                />
              </Popover.Trigger>

              <Popover.Content
                align="center"
                side="right"
                sideOffset={12}
                class="p-1 z-50"
              >
                <ZoneContent
                  {context}
                  availableTimeZones={timeZones}
                  activeTimeZone={timeZone}
                  referencePoint={dateTimeAnchor ??
                    interval?.end ??
                    DateTime.now()}
                  onSelectTimeZone={(z) => {
                    onSelectZone(z);
                    open = false;
                    timeZonePickerOpen = false;
                  }}
                />
              </Popover.Content>
            </Popover.Root>
          </div>
        {/if}

        {#if showTimeDimensionSelector && timeDimensions.length > 1}
          <div class="w-full h-fit px-1">
            <div class="h-px w-full bg-border my-1"></div>

            <Popover.Root bind:open={timeAxisPickerOpen}>
              <Popover.Trigger
                id="time-axis-trigger-{context}"
                onclick={() => {
                  showCalendarPicker = false;
                }}
                aria-label={m.dashboard_select_time_axis()}
                class="group h-7 overflow-hidden hover:bg-surface-hover flex-none rounded-sm w-full select-none flex items-center truncate text-left gap-x-1 pr-1 pl-2"
              >
                <div class="flex-none">
                  <Clock size="14px" />
                </div>
                <div class="mr-auto">{m.dashboard_time_axis()}</div>
                {#if activeTimeDimension}
                  <div class="sr-only group-hover:not-sr-only">
                    <SyntaxElement range={activeTimeDimension.label} />
                  </div>
                {/if}
                <CaretDownIcon className="-rotate-90" size="14px" />
              </Popover.Trigger>

              <Popover.Content
                align="center"
                side="right"
                sideOffset={12}
                class="p-1 z-50"
              >
                {#each timeDimensions as { value, label, description } (value)}
                  <Tooltip.Root>
                    <Tooltip.Trigger>
                      {#snippet child({ props })}
                        <button
                          {...props}
                          class="item"
                          aria-label={m.dashboard_select_time_dimension({
                            label,
                          })}
                          onclick={() => {
                            onSelectTimeDimension(value);
                            open = false;
                            timeAxisPickerOpen = false;
                          }}
                        >
                          {label}
                          {#if value === (timeDimension || primaryTimeDimension)}
                            <Check class="size-4" color="var(--fg-primary)" />
                          {/if}
                        </button>
                      {/snippet}
                    </Tooltip.Trigger>
                    {#if description}
                      <Tooltip.Content
                        side="right"
                        sideOffset={8}
                        class="max-w-64"
                      >
                        <div class="font-bold text-fg-inverse">{label}</div>
                        <div class="text-fg-inverse/70">{description}</div>
                      </Tooltip.Content>
                    {/if}
                  </Tooltip.Root>
                {/each}
              </Popover.Content>
            </Popover.Root>
          </div>
        {/if}
      </div>

      {#if showCalendarPicker}
        <div class="bg-surface-overlay border-l p-3 size-full overflow-y-auto">
          <CalendarPlusDateInput
            {interval}
            zone={timeZone}
            minTimeGrain={V1TimeGrainToDateTimeUnit[
              smallestTimeGrain ?? V1TimeGrain.TIME_GRAIN_MINUTE
            ]}
            {minDate}
            {maxDate}
            {maxQueryTimeRange}
            onApply={() => {
              if (searchValue) onSelectRange(searchValue);
            }}
            updateRange={(string) => {
              searchValue = string;
            }}
            closeMenu={() => (open = false)}
          />
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>

{#if dateTimeAnchor && !hideTruncationSelector}
  <TruncationSelector
    {dateTimeAnchor}
    grain={truncationGrain}
    rangeGrain={parsedTime?.rangeGrain ?? truncationGrain}
    isPeriodToDate={parsedTime?.interval instanceof RillPeriodToGrainInterval}
    {watermark}
    latest={maxDate}
    {smallestTimeGrain}
    {snapToEnd}
    {ref}
    zone={timeZone}
    onSelectEnding={onSelectGrain}
    onToggleAlignment={(inclusive) => {
      onSelectAsOfOption(ref, inclusive);
    }}
    onSelectAsOfOption={(o) => {
      onSelectAsOfOption(o, snapToEnd);
    }}
  />
{/if}

<style lang="postcss">
  .item {
    @apply w-full relative justify-between flex cursor-pointer select-none items-start rounded-sm py-1.5 px-2 gap-x-2 text-xs outline-none;
  }

  .item:hover {
    @apply bg-surface-hover text-fg-primary;
  }
</style>
