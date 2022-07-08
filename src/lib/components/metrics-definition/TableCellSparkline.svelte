<script lang="ts">
  import TimestampSpark from "$lib/components/data-graphic/compositions/timestamp-profile/TimestampSpark.svelte";
  import { convertTimestampPreview } from "$lib/util/convertTimestampPreview";
  import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";
  import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";
  import { store } from "$lib/redux-store/store-root";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import { Debounce } from "$common/utils/Debounce";
  import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
  import type { Readable } from "svelte/store";
  import { getMeasureById } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import { getTimeSeriesById } from "$lib/redux-store/timeseries/timeseries-readables";
  import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

  export let measureId;

  let measure: Readable<MeasureDefinitionEntity>;
  $: measure = getMeasureById(measureId);
  let expression: string;
  let expressionIsValid: ValidationState;
  $: if ($measure) {
    expression = $measure.expression;
    expressionIsValid = $measure.expressionIsValid;
  }

  // FIXME: all of this is app state logic that should be handled
  // in the redux-store. This component should be able to simply
  // read the data from the store by id .
  const debounce = new Debounce();
  function generateSparkLine() {
    debounce.debounce(
      measureId,
      () => {
        store.dispatch(
          generateTimeSeriesApi({
            id: $measure.id,
            measures: [$measure],
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

  let timeSeries: Readable<TimeSeriesEntity>;
  $: if (measureId) {
    timeSeries = getTimeSeriesById(measureId);
  }
</script>

<td class="py-2 px-4 border border-gray-200 hover:bg-gray-200">
  {#if $timeSeries?.spark}
    <TimestampSpark
      data={convertTimestampPreview($timeSeries.spark)}
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
</td>
