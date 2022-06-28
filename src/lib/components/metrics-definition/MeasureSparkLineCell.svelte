<script lang="ts">
  import { ColumnConfig } from "$lib/components/table/pinnableUtils";
  import TimestampSpark from "$lib/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview.js";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config.js";
  import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import { selectTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-selectors";
  import { reduxReadable } from "$lib/redux-store/store-root";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { selectMeasureById } from "$lib/redux-store/measure-definition/measure-definition-selectors";
  import { selectMetricsDefinitionById } from "$lib/redux-store/metrics-definition/metrics-definitioin-selectors";
  import {
    MetricsDefinitionEntity,
    ValidationState,
  } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

  export let value;
  export let index;
  export let column: ColumnConfig;
  export let isNull = false;

  let measure: MeasureDefinitionEntity;
  $: if (value) measure = selectMeasureById(value)($reduxReadable);

  let metricsDef: MetricsDefinitionEntity;
  $: if (measure?.metricsDefId)
    metricsDef = selectMetricsDefinitionById(measure.metricsDefId)(
      $reduxReadable
    );

  $: if (
    measure?.expressionIsValid === ValidationState.OK &&
    metricsDef?.sourceModelId &&
    metricsDef?.timeDimension
  ) {
    console.log(
      measure.expression,
      metricsDef.sourceModelId,
      metricsDef.timeDimension
    );
  }

  let timeSeries: TimeSeriesEntity;
  $: if (value) timeSeries = selectTimeSeriesById(value)($reduxReadable);
</script>

{#if timeSeries?.spark}
  <TimestampSpark
    data={convertTimestampPreview(timeSeries.spark)}
    xAccessor="ts"
    yAccessor="count"
    width={COLUMN_PROFILE_CONFIG.summaryVizWidth.medium}
    height={32}
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
