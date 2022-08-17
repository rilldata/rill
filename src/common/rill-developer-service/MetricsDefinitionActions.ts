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
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { shallowCopy } from "$common/utils/shallowCopy";
import { ExplorerSourceModelDoesntExist } from "$common/errors/ErrorMessages";

export type MetricsDefinitionContext = RillRequestContext<
  EntityType.MetricsDefinition,
  StateType.Persistent
>;

export class MetricsDefinitionActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async createMetricsDefinition(
    rillRequestContext: MetricsDefinitionContext,
    initialFields?: Partial<MetricsDefinitionEntity>
  ) {
    const metricsDefinition = getMetricsDefinition(
      rillRequestContext.entityStateService.getCurrentState().counter
    );
    if (initialFields) {
      delete initialFields.id;
      shallowCopy(initialFields, metricsDefinition);
    }

    this.dataModelerStateService.dispatch(
      "incrementMetricsDefinitionCounter",
      []
    );
    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.MetricsDefinition,
      StateType.Persistent,
      metricsDefinition,
    ]);
    return ActionResponseFactory.getSuccessResponse("", metricsDefinition);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMetricsDefinition(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    modifications: MetricsDefinitionEntity
  ) {
    // TODO: validate ids
    modifications.id = metricsDefId;
    this.dataModelerStateService.dispatch("updateEntity", [
      EntityType.MetricsDefinition,
      StateType.Persistent,
      modifications,
    ]);

    return ActionResponseFactory.getSuccessResponse(
      "",
      this.dataModelerStateService
        .getMetricsDefinitionService()
        .getById(metricsDefId)
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async clearMeasuresAndDimensions(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    this.dataModelerStateService
      .getMeasureDefinitionService()
      .getCurrentState()
      .entities.filter((measure) => measure.metricsDefId === metricsDefId)
      .forEach((measure) => {
        this.dataModelerStateService.dispatch("deleteEntity", [
          EntityType.MeasureDefinition,
          StateType.Persistent,
          measure.id,
        ]);
      });

    this.dataModelerStateService
      .getDimensionDefinitionService()
      .getCurrentState()
      .entities.filter((dimension) => dimension.metricsDefId === metricsDefId)
      .forEach((dimension) => {
        this.dataModelerStateService.dispatch("deleteEntity", [
          EntityType.DimensionDefinition,
          StateType.Persistent,
          dimension.id,
        ]);
      });
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async generateMeasuresAndDimensions(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    if (!rillRequestContext.record || !rillRequestContext.record.sourceModelId)
      return;

    await this.rillDeveloperService.dispatch(
      rillRequestContext,
      "clearMeasuresAndDimensions",
      [metricsDefId]
    );

    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(rillRequestContext.record.sourceModelId);
    if (!model) {
      return ActionResponseFactory.getEntityError(
        ExplorerSourceModelDoesntExist
      );
    }

    await Promise.all(
      model.profile.map((column) =>
        this.inferFromColumn(
          rillRequestContext,
          rillRequestContext.record,
          column
        )
      )
    );
    await this.createCountMeasure(rillRequestContext, metricsDefId);
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
      const resp = await this.rillDeveloperService.dispatch(
        rillRequestContext,
        "addNewDimension",
        [metricsDefinition.id, column.name]
      );
      rillRequestContext.actionsChannel.pushMessage(
        resp.data as Record<string, unknown>
      );
    }
  }

  private async createCountMeasure(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    const measureResp = await this.rillDeveloperService.dispatch(
      rillRequestContext,
      "addNewMeasure",
      [metricsDefId, "count(*)"]
    );
    rillRequestContext.actionsChannel.pushMessage(
      measureResp.data as Record<string, unknown>
    );
  }
}
