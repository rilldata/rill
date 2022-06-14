import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { getMetricsDefinition } from "$common/stateInstancesFactory";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { ProfileColumn } from "$lib/types";
import { CATEGORICALS, NUMERICS } from "$lib/duckdb-data-types";

export type MetricsDefinitionContext = RillRequestContext<
  EntityType.MetricsDefinition,
  StateType.Persistent
>;

export class MetricsDefinitionActions extends RillDeveloperActions {
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
  }

  public async updateMetricsDefinitionTime(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    timeDimension: string
  ) {
    // TODO: validate ids
    this.dataModelerStateService.dispatch("updateMetricsDefinitionTime", [
      metricsDefId,
      timeDimension,
    ]);
  }

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
        this.inferFromColumn(metricsDefinition, column)
      )
    );
  }

  private async inferFromColumn(
    metricsDefinition: MetricsDefinitionEntity,
    column: ProfileColumn
  ) {
    if (CATEGORICALS.has(column.type)) {
      await this.rillDeveloperService.dispatch("addNewDimension", [
        metricsDefinition.id,
        column.name,
      ]);
    } else if (NUMERICS.has(column.type)) {
      // TODO: it is not possible without SQL parsing
    }
  }
}
