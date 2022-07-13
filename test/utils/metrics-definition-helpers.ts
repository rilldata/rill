import type { InlineTestServer } from "./InlineTestServer";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { TestServer } from "./TestServer";
import type {
  BasicMeasureDefinition,
  MeasureDefinitionEntity,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { useTestModel, useTestTables } from "./useInlineTestServer";

/**
 * Call this at the top level to create a metrics definition for a given label, modelName and timeDimension.
 * Make sure to call {@link useTestModel} for the model before this. Also make sure timeDimension passed exists in the model.
 *
 * 1. This creates a MetricsDefinition with given label.
 * 2. Updates sourceModelId and timeDimension for the MetricsDefinition.
 * 3. Calls 'generateMeasuresAndDimensions' to populate measures and dimensions.
 */
export function useMetricsDefinition(
  server: InlineTestServer,
  metricDefLabel: string,
  modelName: string,
  timeDimension: string
) {
  beforeAll(async () => {
    const [model] = server.getModels("tableName", modelName);
    const metricsDef = (
      await server.rillDeveloperService.dispatch(
        RillRequestContext.getNewContext(),
        "createMetricsDefinition",
        []
      )
    ).data as MetricsDefinitionEntity;
    await server.rillDeveloperService.dispatch(
      RillRequestContext.getNewContext(),
      "updateMetricsDefinition",
      [
        metricsDef.id,
        { metricDefLabel, sourceModelId: model.id, timeDimension } as any,
      ]
    );
    await server.rillDeveloperService.dispatch(
      RillRequestContext.getNewContext(),
      "generateMeasuresAndDimensions",
      [metricsDef.id]
    );
  });
}

/**
 * Given a label it selects the MetricsDefinition and its measures and dimensions.
 * These are passed to the callback, use this callback to assign to global variables in the test.
 * Make sure to call this after {@link useMetricsDefinition}.
 */
export function getMetricsDefinition(
  server: TestServer,
  label: string,
  callback: (
    metricsDef: MetricsDefinitionEntity,
    measures: Array<MeasureDefinitionEntity>,
    dimensions: Array<DimensionDefinitionEntity>
  ) => void
) {
  beforeAll(() => {
    const metricsDef = server.getMetricsDefinition("metricDefLabel", label);
    const measures = server.dataModelerStateService
      .getMeasureDefinitionService()
      .getCurrentState()
      .entities.filter((entity) => entity.metricsDefId === metricsDef.id);
    const dimensions = server.dataModelerStateService
      .getDimensionDefinitionService()
      .getCurrentState()
      .entities.filter((entity) => entity.metricsDefId === metricsDef.id);
    callback(metricsDef, measures, dimensions);
  });
}

/**
 * Given a label and measures,
 * 1. This updates count(*) measure name given by 'countMeasureName' argument.
 * 2. Create measures defined by 'otherMeasures' argument.
 * Make sure to call this after {@link useMetricsDefinition}.
 */
export function setupMeasures(
  server: InlineTestServer,
  label: string,
  countMeasureName: string,
  otherMeasures: Array<BasicMeasureDefinition>
) {
  beforeAll(async () => {
    const metricsDef = server.getMetricsDefinition("metricDefLabel", label);
    const countMeasure = server.dataModelerStateService
      .getMeasureDefinitionService()
      .getByField("metricsDefId", metricsDef.id);

    await server.rillDeveloperService.dispatch(
      RillRequestContext.getNewContext(),
      "updateMeasure",
      [countMeasure.id, { sqlName: countMeasureName } as any]
    );

    for (const otherMeasure of otherMeasures) {
      const otherMeasureEntity = (
        await server.rillDeveloperService.dispatch(
          RillRequestContext.getNewContext(),
          "addNewMeasure",
          [metricsDef.id]
        )
      ).data as MeasureDefinitionEntity;
      await server.rillDeveloperService.dispatch(
        RillRequestContext.getNewContext(),
        "updateMeasure",
        [
          otherMeasureEntity.id,
          {
            sqlName: otherMeasure.sqlName,
            expression: otherMeasure.expression,
          } as any,
        ]
      );
    }
  });
}

export function useBasicMetricsDefinition(
  inlineServer: InlineTestServer,

  callback: (
    metricsDef: MetricsDefinitionEntity,
    measures: Array<MeasureDefinitionEntity>,
    dimensions: Array<DimensionDefinitionEntity>
  ) => void
) {
  const AdEventsName = "AdEvents";

  useTestTables(inlineServer);
  useTestModel(
    inlineServer,
    `select
    bid.*, imp.user_id, imp.city, imp.country
    from AdBids bid join AdImpressions imp on bid.id = imp.id`,
    AdEventsName
  );
  useMetricsDefinition(inlineServer, AdEventsName, AdEventsName, "timestamp");
  setupMeasures(inlineServer, AdEventsName, "impressions", [
    { id: "", expression: "avg(bid_price)", sqlName: "bid_price" },
  ]);
  getMetricsDefinition(inlineServer, AdEventsName, callback);
}
