import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import { parseExpression } from "$common/utils/parseQuery";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { guidGenerator } from "$lib/util/guid";

export class MeasuresActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async addNewMeasure(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    const measure: MeasureDefinition = {
      expression: "",
      sqlName: "",
      id: guidGenerator(),
      expressionIsValid: ValidationState.ERROR,
      sqlNameIsValid: ValidationState.ERROR,
      sparkLineId: "",
    };

    this.dataModelerStateService.dispatch("addNewMeasure", [
      metricsDefId,
      measure,
    ]);
    rillRequestContext.actionsChannel.pushMessage("addNewMeasure", [
      metricsDefId,
      measure,
    ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMeasureExpression(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    expression: string
  ) {
    // TODO: validations
    const parsedExpression = parseExpression(expression);
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(rillRequestContext.record.sourceModelId);

    const expressionIsValid =
      parsedExpression.isValid &&
      parsedExpression.columns.every(
        (columnName) =>
          model.profile.findIndex((column) => column.name === columnName) >= 0
      );

    this.dataModelerStateService.dispatch("updateMeasure", [
      metricsDefId,
      measureId,
      {
        expression,
        expressionIsValid: expressionIsValid
          ? ValidationState.OK
          : ValidationState.ERROR,
      },
    ]);
    rillRequestContext.actionsChannel.pushMessage("updateMeasure", [
      metricsDefId,
      measureId,
      {
        expression,
        expressionIsValid: expressionIsValid
          ? ValidationState.OK
          : ValidationState.ERROR,
      },
    ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMeasureSqlName(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    sqlName: string
  ) {
    // TODO: validations
    const modifications: Partial<MeasureDefinition> = {
      sqlName,
      sqlNameIsValid:
        sqlName !== "" ? ValidationState.OK : ValidationState.ERROR,
    };
    this.dataModelerStateService.dispatch("updateMeasure", [
      metricsDefId,
      measureId,
      modifications,
    ]);
    rillRequestContext.actionsChannel.pushMessage("updateMeasure", [
      metricsDefId,
      measureId,
      modifications,
    ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMeasure(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    modifications: Partial<MeasureDefinition>
  ) {
    this.dataModelerStateService.dispatch("updateMeasure", [
      metricsDefId,
      measureId,
      modifications,
    ]);
    rillRequestContext.actionsChannel.pushMessage("updateMeasure", [
      metricsDefId,
      measureId,
      modifications,
    ]);
  }
}
