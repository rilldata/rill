import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { CategoricalSummary } from "$lib/types";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { getDimensionDefinition } from "$common/stateInstancesFactory";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";

/**
 * select
 * count(*), date_trunc('HOUR', created_date) as inter
 * from nyc311_reduced
 * group by date_trunc('HOUR', created_date) order by inter;
 */

const HIGH_CARDINALITY_THRESHOLD = 100;

export class DimensionsActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async addNewDimension(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    columnName?: string
  ) {
    const dimension = getDimensionDefinition(metricsDefId);
    dimension.dimensionColumn = columnName ?? "";

    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.DimensionDefinition,
      StateType.Persistent,
      dimension,
    ]);

    return ActionResponseFactory.getSuccessResponse("", dimension);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateDimension(
    rillRequestContext: MetricsDefinitionContext,
    dimensionId: string,
    modifications: DimensionDefinitionEntity
  ) {
    modifications.id = dimensionId;
    this.dataModelerStateService.dispatch("updateEntity", [
      EntityType.DimensionDefinition,
      StateType.Persistent,
      modifications,
    ]);

    return ActionResponseFactory.getSuccessResponse(
      "",
      this.dataModelerStateService
        .getDimensionDefinitionService()
        .getById(dimensionId)
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async deleteDimension(
    rillRequestContext: MetricsDefinitionContext,
    dimensionId: string
  ) {
    this.dataModelerStateService.dispatch("deleteEntity", [
      EntityType.DimensionDefinition,
      StateType.Persistent,
      dimensionId,
    ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async collectDimensionsInfo(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    // await Promise.all(
    //   rillRequestContext.record.dimensions.map((dimension) =>
    //     this.collectDimensionSummary(
    //       rillRequestContext,
    //       metricsDefId,
    //       dimension.id,
    //       dimension.dimensionColumn,
    //       model.tableName
    //     )
    //   )
    // );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateDimensionColumn(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    dimensionId: string,
    columnName: string
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const derivedModel = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(rillRequestContext.record.sourceModelId);

    const modifications: Partial<DimensionDefinitionEntity> = {
      dimensionColumn: columnName,
      dimensionIsValid:
        derivedModel.profile.findIndex(
          (column) => column.name === columnName
        ) >= 0
          ? ValidationState.OK
          : ValidationState.ERROR,
    };
    // this.dataModelerStateService.dispatch("updateDimension", [
    //   metricsDefId,
    //   dimensionId,
    //   modifications,
    // ]);
    // rillRequestContext.actionsChannel.pushMessage("updateDimension", [
    //   metricsDefId,
    //   dimensionId,
    //   modifications,
    // ]);

    if (modifications.dimensionIsValid) {
      await this.collectDimensionSummary(
        rillRequestContext,
        metricsDefId,
        dimensionId,
        columnName,
        model.tableName
      );
    }
  }

  private async collectDimensionSummary(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    dimensionId: string,
    dimensionColumn: string,
    tableName: string
  ) {
    const summary: CategoricalSummary = await this.databaseActionQueue.enqueue(
      { id: metricsDefId, priority: DatabaseActionQueuePriority.ActiveModel },
      "getTopKAndCardinality",
      [tableName, dimensionColumn]
    );
    const modifications: Partial<DimensionDefinitionEntity> = {
      summary,
      dimensionIsValid:
        summary.cardinality >= HIGH_CARDINALITY_THRESHOLD
          ? ValidationState.WARNING
          : ValidationState.OK,
    };
    // this.dataModelerStateService.dispatch("updateDimension", [
    //   metricsDefId,
    //   dimensionId,
    //   modifications,
    // ]);
    // rillRequestContext.actionsChannel.pushMessage("updateDimension", [
    //   metricsDefId,
    //   dimensionId,
    //   modifications,
    // ]);
  }
}
