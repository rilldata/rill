import { RillDeveloperActions } from "./RillDeveloperActions";
import type { MetricsDefinitionContext } from "./MetricsDefinitionActions";
import { ValidationState } from "../data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { DimensionDefinitionEntity } from "../data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { getDimensionDefinition } from "../stateInstancesFactory";
import { ActionResponseFactory } from "../data-modeler-service/response/ActionResponseFactory";
import { ExplorerSourceModelDoesntExist } from "../errors/ErrorMessages";

/**
 * select
 * count(*), date_trunc('HOUR', created_date) as inter
 * from nyc311_reduced
 * group by date_trunc('HOUR', created_date) order by inter;
 */

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
  public async validateDimensionColumn(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    columnName: string
  ) {
    if (!metricsDefId || !rillRequestContext.record)
      return ActionResponseFactory.getEntityError(
        `No metrics found for id=${metricsDefId}`
      );

    const derivedModel = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(rillRequestContext.record.sourceModelId);
    if (!derivedModel) {
      return ActionResponseFactory.getEntityError(
        ExplorerSourceModelDoesntExist
      );
    }

    const columnFindIndex = derivedModel.profile.findIndex(
      (column) => column.name === columnName
    );
    return ActionResponseFactory.getSuccessResponse("", {
      dimensionIsValid:
        columnFindIndex >= 0 ? ValidationState.OK : ValidationState.ERROR,
    });
  }
}
