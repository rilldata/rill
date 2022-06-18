import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { DimensionDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { CategoricalSummary } from "$lib/types";
import { guidGenerator } from "$lib/util/guid";

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
    columnName: string
  ) {
    const dimensions: DimensionDefinition = {
      dimensionColumn: columnName,
      id: guidGenerator(),
      dimensionIsValid:
        columnName === "" ? ValidationState.OK : ValidationState.ERROR,
      sqlNameIsValid: ValidationState.OK,
    };

    this.dataModelerStateService.dispatch("addNewDimension", [
      metricsDefId,
      dimensions,
    ]);
    rillRequestContext.actionsChannel.pushMessage("addNewDimension", [
      metricsDefId,
      dimensions,
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
    await Promise.all(
      rillRequestContext.record.dimensions.map((dimension) =>
        this.collectDimensionSummary(
          rillRequestContext,
          metricsDefId,
          dimension.id,
          dimension.dimensionColumn,
          model.tableName
        )
      )
    );
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

    const modifications: Partial<DimensionDefinition> = {
      dimensionColumn: columnName,
      dimensionIsValid:
        derivedModel.profile.findIndex(
          (column) => column.name === columnName
        ) >= 0
          ? ValidationState.OK
          : ValidationState.ERROR,
    };
    this.dataModelerStateService.dispatch("updateDimension", [
      metricsDefId,
      dimensionId,
      modifications,
    ]);
    rillRequestContext.actionsChannel.pushMessage("updateDimension", [
      metricsDefId,
      dimensionId,
      modifications,
    ]);

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

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateDimension(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    dimensionId: string,
    modifications: Partial<DimensionDefinition>
  ) {
    this.dataModelerStateService.dispatch("updateDimension", [
      metricsDefId,
      dimensionId,
      modifications,
    ]);
    rillRequestContext.actionsChannel.pushMessage("updateDimension", [
      metricsDefId,
      dimensionId,
      modifications,
    ]);
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
    const modifications: Partial<DimensionDefinition> = {
      summary,
      dimensionIsValid:
        summary.cardinality >= HIGH_CARDINALITY_THRESHOLD
          ? ValidationState.WARNING
          : ValidationState.OK,
    };
    this.dataModelerStateService.dispatch("updateDimension", [
      metricsDefId,
      dimensionId,
      modifications,
    ]);
    rillRequestContext.actionsChannel.pushMessage("updateDimension", [
      metricsDefId,
      dimensionId,
      modifications,
    ]);
  }
}
