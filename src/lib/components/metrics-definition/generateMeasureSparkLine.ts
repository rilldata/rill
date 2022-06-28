import { Debounce } from "$common/utils/Debounce";
import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
import { store } from "$lib/redux-store/store-root";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { COLUMN_PROFILE_CONFIG } from "$lib/application-config";

const debounce = new Debounce();

export function generateMeasureSparkLine(
  metricsDefId: string,
  measures: Array<MeasureDefinitionEntity>
) {
  debounce.debounce(
    metricsDefId,
    () => {
      store.dispatch(
        generateTimeSeriesApi({
          metricsDefId,
          measures,
          filters: {},
          pixels: COLUMN_PROFILE_CONFIG.summaryVizWidth.medium,
        })
      );
    },
    1000
  );
}
