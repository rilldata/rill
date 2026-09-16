<script lang="ts">
  import { getComparisonLabelFromRange } from "@rilldata/web-common/lib/time/comparisons/index.ts";
  import RangeDisplay from "@rilldata/web-common/features/dashboards/time-controls/super-pill/components/RangeDisplay.svelte";
  import { Chip } from "@rilldata/web-common/components/chip/index.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import { RillIsoInterval } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";

  let {
    timeFilterManager,
    hasBoldTimeRange = true,
  }: {
    timeFilterManager: TimeFilterManager;
    hasBoldTimeRange?: boolean;
  } = $props();

  let {
    timeRange,
    timeGrain,
    interval,
    parsedTime,
    comparisonTimeRange,
    parsedComparisonTime,
  } = $derived(timeFilterManager);

  let selectedLabel = $derived(getRangeLabel(timeRange));
  let showRange = $derived(
    parsedTime && parsedTime?.interval instanceof RillIsoInterval,
  );
</script>

<Chip type="time" theme readOnly>
  <svelte:fragment slot="body">
    <div class="text-fg-primary flex gap-x-1.5">
      <div class="font-bold">
        {#if showRange}
          {m.time_custom()}
        {:else}
          {selectedLabel}
        {/if}
      </div>
      {#if showRange && interval}
        <RangeDisplay {interval} {timeGrain} />
      {/if}
    </div>
  </svelte:fragment>
</Chip>

{#if comparisonTimeRange && parsedComparisonTime}
  <Chip type="time" readOnly>
    <svelte:fragment slot="body">
      <div class="text-fg-primary px-2">
        {m.time_vs()}
        <span class:font-bold={hasBoldTimeRange}>
          {getComparisonLabelFromRange(
            comparisonTimeRange,
            parsedComparisonTime,
          )}
        </span>
      </div>
    </svelte:fragment>
  </Chip>
{/if}
