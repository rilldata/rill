import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import { parseExpression } from "$common/utils/parseQuery";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { getMeasureDefinition } from "$common/stateInstancesFactory";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export class MeasuresActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async addNewMeasure(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    expression?: string
  ) {
    const measure = getMeasureDefinition(metricsDefId);
    if (expression) {
      measure.expression = expression;
    }

    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.MeasureDefinition,
      StateType.Persistent,
      measure,
    ]);

    const newMeasure = { ...measure };
    if (expression) {
      const expressionValidationResp = await this.rillDeveloperService.dispatch(
        rillRequestContext,
        "validateMeasureExpression",
        [metricsDefId, expression]
      );
      newMeasure.expressionIsValid = (
        expressionValidationResp?.data as any
      ).expressionIsValid;
    }

    return ActionResponseFactory.getSuccessResponse("", newMeasure);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMeasure(
    rillRequestContext: MetricsDefinitionContext,
    measureId: string,
    modifications: MeasureDefinitionEntity
  ) {
    modifications.id = measureId;
    this.dataModelerStateService.dispatch("updateEntity", [
      EntityType.MeasureDefinition,
      StateType.Persistent,
      modifications,
    ]);
    return ActionResponseFactory.getSuccessResponse(
      "",
      this.dataModelerStateService
        .getMeasureDefinitionService()
        .getById(measureId)
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async deleteMeasure(
    rillRequestContext: MetricsDefinitionContext,
    measureId: string
  ) {
    this.dataModelerStateService.dispatch("deleteEntity", [
      EntityType.MeasureDefinition,
      StateType.Persistent,
      measureId,
    ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async validateMeasureExpression(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    expression: string
  ) {
    if (!metricsDefId || !rillRequestContext.record)
      return ActionResponseFactory.getEntityError(
        `No metrics found for id=${metricsDefId}`
      );

    const parsedExpression = parseExpression(expression);
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(rillRequestContext.record.sourceModelId);

    const expressionIsValid =
      parsedExpression.isValid &&
      parsedExpression.columns.every(
        (columnName) =>
          columnName === "*" ||
          model.profile.findIndex((column) => column.name === columnName) >= 0
      );

    return ActionResponseFactory.getSuccessResponse("", {
      expressionIsValid: expressionIsValid
        ? ValidationState.OK
        : ValidationState.ERROR,
    });
  }
}
