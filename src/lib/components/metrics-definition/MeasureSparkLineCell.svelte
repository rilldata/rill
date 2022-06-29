<script lang="ts">
  import TimestampSpark from "$lib/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import { selectTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-selectors";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { selectMeasureById } from "$lib/redux-store/measure-definition/measure-definition-selectors";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { Debounce } from "$common/utils/Debounce";
  import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
  import type { ColumnConfig } from "$lib/components/table/ColumnConfig";

  export let value;
  export let index;
  export let column: ColumnConfig;
  export let isNull = false;

  let measure: MeasureDefinitionEntity;
  $: if (value) measure = selectMeasureById(value)($reduxReadable);
  let metricsDefId: string;
  let expression: string;
  let expressionIsValid: ValidationState;
  $: if (measure) {
    metricsDefId = measure.metricsDefId;
    expression = measure.expression;
    expressionIsValid = measure.expressionIsValid;
  }

  const debounce = new Debounce();
  function generateSparkLine() {
    debounce.debounce(
      value,
      () => {
        store.dispatch(
          generateTimeSeriesApi({
            metricsDefId,
            measures: [measure],
            filters: {},
            pixels: COLUMN_PROFILE_CONFIG.summaryVizWidth.medium,
          })
        );
      },
      1000
    );
  }
  $: if (expression && expressionIsValid === ValidationState.OK) {
    generateSparkLine();
  }

  let timeSeries: TimeSeriesEntity;
  $: if (value) {
    timeSeries = selectTimeSeriesById(value)($reduxReadable);
  }
</script>

{#if timeSeries?.spark}
  <TimestampSpark
    data={convertTimestampPreview(timeSeries.spark)}
    xAccessor="ts"
    yAccessor="count"
    width={COLUMN_PROFILE_CONFIG.summaryVizWidth.medium}
    height={20}
    top={0}
    bottom={0}
    left={0}
    right={0}
    leftBuffer={0}
    rightBuffer={0}
    area
    tweenIn
  />
{/if}
