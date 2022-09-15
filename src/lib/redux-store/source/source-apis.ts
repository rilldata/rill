import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { dataModelerService } from "$lib/application-state-stores/application-store";
import {
  resetQuickStartDashboardOverlay,
  showQuickStartDashboardOverlay,
} from "$lib/application-state-stores/layout-store";
import notificationStore from "$lib/components/notifications";
import { TIMESTAMPS } from "$lib/duckdb-data-types";
import {
  createMetricsDefsApi,
  generateMeasuresAndDimensionsApi,
} from "$lib/redux-store/metrics-definition/metrics-definition-apis";
import { selectMetricsDefinitionMatchingName } from "$lib/redux-store/metrics-definition/metrics-definition-selectors";
import { updateModelQueryApi } from "$lib/redux-store/model/model-apis";
import {
  selectDerivedModelBySourceName,
  selectPersistentModelById,
} from "$lib/redux-store/model/model-selector";
import { store } from "$lib/redux-store/store-root";

// Source doesn't have a slice as of now.
// This file has simple code that will eventually be moved into async thunks

export const deleteSourceApi = async (persistentTableName: string) => {
  await dataModelerService.dispatch("dropTable", [persistentTableName]);
  await sourceUpdated(persistentTableName);
};

/**
 * Called when a source is created or deleted.
 */
export const sourceUpdated = async (persistentTableName: string) => {
  await Promise.all(
    selectDerivedModelBySourceName(persistentTableName).map((derivedModel) =>
      updateModelQueryApi(
        derivedModel.id,
        selectPersistentModelById(derivedModel.id).query,
        true
      )
    )
  );
};

/**
 * Create a model for the given source by selecting all columns.
 */
export const createModelForSource = async (
  models: Array<PersistentModelEntity>,
  sourceName: string
) => {
  const createdModelId = await createModelFromSourceAndGetId(
    models,
    sourceName
  );
  // change the active asset to the new model
  await dataModelerService.dispatch("setActiveAsset", [
    EntityType.Model,
    createdModelId,
  ]);

  notificationStore.send({
    message: `queried ${sourceName} in workspace`,
  });

  return createdModelId;
};

/**
 * Quick starts a metrics dashboard for a given source.
 * The source should have a timestamp column for this to work.
 */
export const autoCreateMetricsDefinitionForSource = async (
  models: Array<PersistentModelEntity>,
  derivedSources: Array<DerivedTableEntity>,
  id: string,
  sourceName: string
) => {
  let createdMetricsId: string = null;
  try {
    const timestampColumns = derivedSources
      .find((source) => source.id === id)
      .profile?.filter((column) => TIMESTAMPS.has(column.type));
    if (!timestampColumns?.length) return;
    showQuickStartDashboardOverlay(sourceName, timestampColumns[0].name);
    const modelId = await createModelFromSourceAndGetId(models, sourceName);

    createdMetricsId = await autoCreateMetricsDefinitionForModel(
      sourceName,
      modelId,
      timestampColumns[0].name
    );
  } catch (e) {
    console.error(e);
  }
  resetQuickStartDashboardOverlay();
  return createdMetricsId;
};

/**
 * Creates a metrics definition for a given model, time dimension and a label.
 * Auto generates measures and dimensions.
 * Focuses the dashboard created.
 */
export const autoCreateMetricsDefinitionForModel = async (
  sourceName: string,
  sourceModelId: string,
  timeDimension: string
): Promise<string> => {
  const metricsLabel = `${sourceName}_dashboard`;
  const existingMetrics = selectMetricsDefinitionMatchingName(
    store.getState(),
    metricsLabel
  );

  const { payload: createdMetricsDef } = await store.dispatch(
    createMetricsDefsApi({
      sourceModelId,
      timeDimension,
      metricDefLabel:
        existingMetrics.length === 0
          ? metricsLabel
          : `${metricsLabel}_${existingMetrics.length}`,
    })
  );

  await store.dispatch(generateMeasuresAndDimensionsApi(createdMetricsDef.id));
  await dataModelerService.dispatch("setActiveAsset", [
    EntityType.MetricsExplorer,
    createdMetricsDef.id,
  ]);

  return createdMetricsDef.id;
};

/**
 * Create a model with all columns from the source
 */
const createModelFromSourceAndGetId = async (
  models: Array<PersistentModelEntity>,
  sourceName: string
): Promise<string> => {
  // check existing models to avoid a name conflict
  const existingNames = models
    .filter((model) => model.name.includes(`${sourceName}_model`))
    .map((model) => model.tableName)
    .sort();
  const nextName =
    existingNames.length === 0
      ? `${sourceName}_model`
      : `${sourceName}_model_${existingNames.length + 1}`;

  const response = await dataModelerService.dispatch("addModel", [
    {
      name: nextName,
      query: `select * from ${sourceName}`,
    },
  ]);
  return (response as unknown as { id: string }).id;
};
