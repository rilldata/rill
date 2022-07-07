<script lang="ts">
  import MetricsExploreTimeChart from "$lib/components/leaderboard/MetricsExploreTimeChart.svelte";
  import type { Readable } from "svelte/store";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getFallbackMeasureName } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import MeasureBigNumber from "$lib/components/leaderboard/MeasureBigNumber.svelte";

  export let metricsDefId: string;
  export let measureId: string;
  export let index: number;

  let measure: Readable<MeasureDefinitionEntity>;
  $: measure = getMeasureById(measureId);
  let measureField: string;
  $: if ($measure) {
    measureField = getFallbackMeasureName(index, $measure.sqlName);
  }
</script>

{#if $measure}
  <div>
    <div class="grid grid grid-flow-col">
      <div class="big-number" style:width="200px">
        <h2>{$measure.label?.length ? $measure.label : measureField}</h2>
        <div><MeasureBigNumber {metricsDefId} {measureId} {index} /></div>
      </div>
      <MetricsExploreTimeChart {metricsDefId} yAccessor={measureField} />
    </div>
  </div>
{/if}
