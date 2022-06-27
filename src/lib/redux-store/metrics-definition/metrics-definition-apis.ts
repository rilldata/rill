import { generateApis } from "$lib/redux-store/slice-utils";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyMetricsDefs,
  addOneMetricsDef,
  removeMetricsDef,
  updateMetricsDef,
} from "$lib/redux-store/metrics-definition/metrics-definition-slice";

export const {
  fetchManyApi: fetchManyMetricsDefsApi,
  createApi: createMetricsDefsApi,
  updateApi: updateMetricsDefsApi,
  deleteApi: deleteMetricsDefsApi,
} = generateApis<EntityType.MetricsDefinition>(
  EntityType.MetricsDefinition,
  addManyMetricsDefs,
  addOneMetricsDef,
  updateMetricsDef,
  removeMetricsDef,
  "metrics"
);
