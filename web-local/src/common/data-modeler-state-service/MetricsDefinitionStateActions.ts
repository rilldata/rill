import { StateActions } from "./StateActions";
import type { MetricsDefinitionStateActionArg } from "./entity-state-service/MetricsDefinitionEntityService";
import type { RollupInterval } from "../database-service/DatabaseColumnActions";

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
