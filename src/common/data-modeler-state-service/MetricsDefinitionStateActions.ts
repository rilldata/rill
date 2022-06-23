import { StateActions } from "$common/data-modeler-state-service/StateActions";
import type { MetricsDefinitionStateActionArg } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";
import { shallowCopy } from "$common/utils/shallowCopy";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

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
}
