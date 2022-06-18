import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getMetricsDefinition } from "$common/stateInstancesFactory";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { ProfileColumn } from "$lib/types";
import { CATEGORICALS } from "$lib/duckdb-data-types";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";

export type MetricsDefinitionContext = RillRequestContext<
  EntityType.MetricsDefinition,
  StateType.Persistent
>;

export class MetricsDefinitionActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async createMetricsDefinition(
    rillRequestContext: MetricsDefinitionContext
  ) {
    const metricsDefinition = getMetricsDefinition(
      rillRequestContext.entityStateService.getCurrentState().counter
    );
    this.dataModelerStateService.dispatch(
      "incrementMetricsDefinitionCounter",
      []
    );
    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.MetricsDefinition,
      StateType.Persistent,
      metricsDefinition,
    ]);
    rillRequestContext.actionsChannel.pushMessage("addEmptyMetricsDef", [
      metricsDefinition,
    ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMetricsDefinitionModel(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    modelId: string
  ) {
    // TODO: validate ids
    this.dataModelerStateService.dispatch("updateMetricsDefinitionModel", [
      metricsDefId,
      modelId,
    ]);
    rillRequestContext.actionsChannel.pushMessage(
      "updateMetricsDefinitionModel",
      [metricsDefId, modelId]
    );
    this.dataModelerStateService.dispatch("clearMetricsDefinition", [
      metricsDefId,
    ]);
    rillRequestContext.actionsChannel.pushMessage("clearMetricsDimension", [
      metricsDefId,
    ]);
    await this.rillDeveloperService.dispatch(
      rillRequestContext,
      "inferMeasuresAndDimensions",
      [metricsDefId]
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMetricsDefinitionTimestamp(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    timeDimension: string
  ) {
    // TODO: validate ids
    this.dataModelerStateService.dispatch("updateMetricsDefinitionTimestamp", [
      metricsDefId,
      timeDimension,
    ]);
    if (!rillRequestContext.record.sourceModelId) return;
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);

    const rollupInterval = await this.databaseActionQueue.enqueue(
      { id: metricsDefId, priority: DatabaseActionQueuePriority.ActiveModel },
      "estimateIdealRollupInterval",
      [model.tableName, timeDimension]
    );
    this.dataModelerStateService.dispatch(
      "updateMetricsDefinitionRollupInterval",
      [metricsDefId, rollupInterval]
    );
    // TODO: update all measure graphs
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async inferMeasuresAndDimensions(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    // TODO: validate ids
    const metricsDefinition =
      rillRequestContext.entityStateService.getById(metricsDefId);
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(metricsDefinition.sourceModelId);

    await Promise.all(
      model.profile.map((column) =>
        this.inferFromColumn(rillRequestContext, metricsDefinition, column)
      )
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async deleteMetricsDefinition(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    this.dataModelerStateService.dispatch("deleteEntity", [
      EntityType.MetricsDefinition,
      StateType.Persistent,
      metricsDefId,
    ]);
  }

  private async inferFromColumn(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefinition: MetricsDefinitionEntity,
    column: ProfileColumn
  ) {
    if (CATEGORICALS.has(column.type)) {
      await this.rillDeveloperService.dispatch(
        rillRequestContext,
        "addNewDimension",
        [metricsDefinition.id, column.name]
      );
    }
  }
}
