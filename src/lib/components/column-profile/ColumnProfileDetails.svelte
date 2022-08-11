<script lang="ts">
  import { afterUpdate } from "svelte";
  import { slide } from "svelte/transition";
  import { dataModelerService } from "$lib/application-state-stores/application-store";

  import TopKSummary from "$lib/components/viz/TopKSummary.svelte";
  import {
    CATEGORICALS,
    NUMERICS,
    TIMESTAMPS,
    DATA_TYPE_COLORS,
  } from "$lib/duckdb-data-types";

  import TimestampHistogram from "$lib/components/viz/histogram/TimestampHistogram.svelte";
  import NumericHistogram from "$lib/components/viz/histogram/NumericHistogram.svelte";
  import OutlierHistogram from "$lib/components/viz/histogram/OutlierHistogram.svelte";
  import { TimestampDetail } from "../data-graphic/compositions/timestamp-profile";

  export let type;
  export let summary;
  export let totalRows;
  export let name;
  export let entityId;
  export let containerWidth: number;

  export let indentLevel = 1;

  export let active = false;

  // Make sure priority is updated in case the profile is already opened
  afterUpdate(async () => {
    if (active) {
      dataModelerService.dispatch("updateColumnProfilePriority", [
        entityId,
        name,
      ]);
    }
  });
  // $: if (active) {
  //   dataModelerService.dispatch("updateColumnProfilePriority", [entityId, name]);
  // }
</script>

<!-- FIXME: document all magic number sums of indent levels in this component,
  and potentially move to another file -->
{#if active}
  <div transition:slide|local={{ duration: 200 }} class="pt-3 pb-3  w-full">
    {#if CATEGORICALS.has(type) && summary?.topK}
      <div class="pl-{indentLevel === 1 ? 16 : 10} pr-4 w-full">
        <!-- pl-16 pl-8 -->
        <TopKSummary
          {containerWidth}
          color={DATA_TYPE_COLORS["VARCHAR"].bgClass}
          {totalRows}
          topK={summary.topK}
        />
      </div>
    {:else if NUMERICS.has(type) && summary?.statistics && summary?.histogram?.length}
      <div class="pl-{indentLevel === 1 ? 12 : 4}">
        <!-- pl-12 pl-5 -->
        <!-- FIXME: we have to remove a bit of pad from the right side to make this work -->
        <NumericHistogram
          width={containerWidth - (indentLevel === 1 ? 20 + 24 + 44 : 32)}
          height={65}
          data={summary.histogram}
          min={summary.statistics.min}
          qlow={summary.statistics.q25}
          median={summary.statistics.q50}
          qhigh={summary.statistics.q75}
          mean={summary.statistics.mean}
          max={summary.statistics.max}
        />
        {#if summary?.outliers && summary?.outliers?.length}
          <OutlierHistogram
            width={containerWidth - (indentLevel === 1 ? 20 + 24 + 44 : 32)}
            height={15}
            data={summary.outliers}
            mean={summary.statistics.mean}
            sd={summary.statistics.sd}
            min={summary.statistics.min}
            max={summary.statistics.max}
          />
        {/if}
      </div>
    {:else if TIMESTAMPS.has(type) && summary?.rollup}
      <div class="pl-{indentLevel === 1 ? 16 : 10}">
        <!-- pl-14 pl-10 -->
        <TimestampDetail
          {type}
          data={summary.rollup.results.map((di) => {
            let pi = { ...di };
            pi.ts = new Date(pi.ts);
            return pi;
          })}
          spark={summary.rollup.spark.map((di) => {
            let pi = { ...di };
            pi.ts = new Date(pi.ts);
            return pi;
          })}
          xAccessor="ts"
          yAccessor="count"
          mouseover={true}
          height={160}
          width={containerWidth - (indentLevel === 1 ? 20 + 24 + 54 : 32 + 20)}
          rollupGrain={summary.rollup.rollupInterval}
          estimatedSmallestTimeGrain={summary?.estimatedSmallestTimeGrain}
          interval={summary.interval}
        />
      </div>
    {:else if TIMESTAMPS.has(type) && summary?.histogram?.length}
      <div class="pl-{indentLevel === 1 ? 16 : 10}">
        <TimestampHistogram
          {type}
          width={containerWidth - (indentLevel === 1 ? 20 + 24 + 54 : 32 + 20)}
          data={summary.histogram}
          interval={summary.interval}
          estimatedSmallestTimeGrain={summary?.estimatedSmallestTimeGrain}
        />
      </div>
    {/if}
  </div>
{/if}
