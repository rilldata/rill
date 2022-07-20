<script lang="ts">
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import { store } from "$lib/redux-store/store-root";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
  import type { Readable } from "svelte/store";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
  import type { EntityRecord } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { CellSparkline } from "$lib/components/table-editable/ColumnConfig";
  import { GraphicContext } from "$lib/components/data-graphic/elements/index";
  import TimeSeriesBody from "../../../routes/_surfaces/workspace/explore/time-series-charts/TimeSeriesBody.svelte";
  import { NicelyFormattedTypes } from "$lib/util/humanize-numbers";

  export let columnConfig: ColumnConfig<CellSparkline>;
  export let index = undefined;
  export let row: EntityRecord;
  $: value = row[columnConfig.name];

  let measure: Readable<MeasureDefinitionEntity>;
  $: measure = getMeasureById(row?.id);
  let expression: string;
  let expressionIsValid: ValidationState;
  $: if ($measure) {
    expression = $measure.expression;
    expressionIsValid = $measure.expressionIsValid;
  }

  function generateSparkLine() {
    store.dispatch(
      generateTimeSeriesApi({
        id: value,
        measures: [$measure],
        filters: {},
        timeRange: {},
        pixels: COLUMN_PROFILE_CONFIG.summaryVizWidth.medium,
        isolated: true,
      })
    );
  }
  $: if (expression && expressionIsValid === ValidationState.OK) {
    generateSparkLine();
  }

  let timeSeries: Readable<TimeSeriesEntity>;
  $: if ($measure?.id) {
    timeSeries = getTimeSeriesById($measure?.id);
  }
</script>

<td class=" border border-gray-200 hover:bg-gray-200">
  {#if $timeSeries?.spark}
    <GraphicContext
      width={300}
      height={25}
      left={24}
      right={45}
      top={4}
      bottom={4}
      xMin={new Date($timeSeries.timeRange.start)}
      xMax={new Date($timeSeries.timeRange.end)}
      yMin={0}
      xType="date"
      yType="number"
      xMinTweenProps={{ duration: 200 }}
      xMaxTweenProps={{ duration: 200 }}
    >
      <TimeSeriesBody
        formatPreset={NicelyFormattedTypes.HUMANIZE}
        data={convertTimestampPreview($timeSeries.spark)}
        key={$measure.id}
        accessor={`measure_${index}`}
      />
    </GraphicContext>
  {/if}
</td>
