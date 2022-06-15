import { StateActions } from "$common/data-modeler-state-service/StateActions";
import type {
  DimensionDefinition,
  MeasureDefinition,
  MetricsDefinitionStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";
import { shallowCopy } from "$common/utils/shallowCopy";

export class MetricsDefinitionStateActions extends StateActions {
  @StateActions.MetricsDefinitionAction()
  public incrementMetricsDefinitionCounter({
    draftState,
  }: MetricsDefinitionStateActionArg) {
    draftState.counter++;
  }

  @StateActions.MetricsDefinitionAction()
  public updateMetricsDefinitionModel(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    modelId: string
  ) {
    stateService.updateEntityField(
      draftState,
      metricsDefId,
      "sourceModelId",
      modelId
    );
  }

  @StateActions.MetricsDefinitionAction()
  public updateMetricsDefinitionTimestamp(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    timeDimension: string
  ) {
    stateService.updateEntityField(
      draftState,
      metricsDefId,
      "timeDimension",
      timeDimension
    );
  }

  @StateActions.MetricsDefinitionAction()
  public updateMetricsDefinitionRollupInterval(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    rollupInterval: RollupInterval
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    metricsDef.rollupInterval = rollupInterval;
  }

  @StateActions.MetricsDefinitionAction()
  public clearMetricsDefinition(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    metricsDef.dimensions = [];
    metricsDef.measures = [];
  }

  @StateActions.MetricsDefinitionAction()
  public addNewDimension(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    dimension: DimensionDefinition
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    metricsDef.dimensions.push(dimension);
  }

  @StateActions.MetricsDefinitionAction()
  public updateDimension(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    dimensionId: string,
    modifications: Partial<DimensionDefinition>
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    const dimensionToUpdate = metricsDef.dimensions.find(
      (dimension) => dimension.id === dimensionId
    );
    shallowCopy(modifications, dimensionToUpdate);
  }

  @StateActions.MetricsDefinitionAction()
  public addNewMeasure(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    measure: MeasureDefinition
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    metricsDef.measures.push(measure);
  }

  @StateActions.MetricsDefinitionAction()
  public updateMeasure(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    measureId: string,
    modifications: Partial<MeasureDefinition>
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    const measureToUpdate = metricsDef.measures.find(
      (measure) => measure.id === measureId
    );
    shallowCopy(modifications, measureToUpdate);
  }
}
