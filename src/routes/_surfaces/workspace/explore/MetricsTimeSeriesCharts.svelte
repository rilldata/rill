<script lang="ts">
  import MetricsExploreTimeChart from "$lib/components/leaderboard/MetricsExploreTimeChart.svelte";
  import type { Readable } from "svelte/store";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";

  export let metricsDefId: string;
  export let measureId: string;
  export let index: number;

  let measure: Readable<MeasureDefinitionEntity>;
  $: measure = getMeasureById(measureId);
</script>

{#if $measure}
  <div>
    <div class="grid grid grid-flow-col">
      <div class="big-number">
        <h2>{$measure.label ?? $measure.sqlName ?? $measure.expression}</h2>
        <div />
      </div>
      <MetricsExploreTimeChart
        {metricsDefId}
        yAccessor={$measure.sqlName ?? `measure_${index}`}
      />
    </div>
  </div>
{/if}
