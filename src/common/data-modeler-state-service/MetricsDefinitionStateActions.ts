import { StateActions } from "$common/data-modeler-state-service/StateActions";
import type {
  DimensionDefinition,
  MetricsDefinitionStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

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
  public updateMetricsDefinitionTime(
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
  public addNewDimension(
    { draftState, stateService }: MetricsDefinitionStateActionArg,
    metricsDefId: string,
    dimension: DimensionDefinition
  ) {
    const metricsDef = stateService.getById(metricsDefId, draftState);
    metricsDef.dimensions.push(dimension);
  }
}
