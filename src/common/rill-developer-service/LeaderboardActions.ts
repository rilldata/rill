import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { ActiveValues } from "$lib/redux-store/metrics-leaderboard-slice";
import type {
  DimensionDefinition,
  MeasureDefinition,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class LeaderboardActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async getLeaderboardValues(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    filters: ActiveValues
  ) {
    const measure = rillRequestContext.record.measures.find(
      (measure) => measure.id === measureId
    );
    await Promise.all(
      rillRequestContext.record.dimensions.map((dimension) =>
        this.getLeaderboardValuesForDimension(
          rillRequestContext,
          measure,
          dimension,
          filters
        )
      )
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getBigNumber(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    filters: ActiveValues
  ) {
    const measure = rillRequestContext.record.measures.find(
      (measure) => measure.id === measureId
    );
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const bigNumberValues = await this.databaseActionQueue.enqueue(
      {
        id: rillRequestContext.id,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "getBigNumber",
      [model.tableName, measure.expression, filters]
    );
    return bigNumberValues.value;
  }

  private async getLeaderboardValuesForDimension(
    rillRequestContext: MetricsDefinitionContext,
    measure: MeasureDefinition,
    dimension: DimensionDefinition,
    filters: ActiveValues
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    rillRequestContext.actionsChannel.pushMessage("setDimensionLeaderboard", [
      rillRequestContext.id,
      dimension.dimensionColumn,
      await this.databaseActionQueue.enqueue(
        {
          id: rillRequestContext.id,
          priority: DatabaseActionQueuePriority.ActiveModel,
        },
        "getLeaderboardValues",
        [
          model.tableName,
          dimension.dimensionColumn,
          measure.expression,
          filters,
        ]
      ),
    ]);
  }
}
