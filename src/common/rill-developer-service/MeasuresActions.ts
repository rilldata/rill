import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import { parseExpression } from "$common/utils/parseQuery";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { getMeasureDefinition } from "$common/stateInstancesFactory";

export class MeasuresActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async addNewMeasure(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    const measure = getMeasureDefinition(metricsDefId);

    this.dataModelerStateService.dispatch("addEntity", [
      EntityType.MeasureDefinition,
      StateType.Persistent,
      measure,
    ]);

    return measure;
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
    return this.dataModelerStateService
      .getMeasureDefinitionService()
      .getById(measureId);
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
          columnName === "*" ||
          model.profile.findIndex((column) => column.name === columnName) >= 0
      );

    // this.dataModelerStateService.dispatch("updateMeasure", [
    //   metricsDefId,
    //   measureId,
    //   {
    //     expression,
    //     expressionIsValid: expressionIsValid
    //       ? ValidationState.OK
    //       : ValidationState.ERROR,
    //   },
    // ]);
    // rillRequestContext.actionsChannel.pushMessage("updateMeasure", [
    //   metricsDefId,
    //   measureId,
    //   {
    //     expression,
    //     expressionIsValid: expressionIsValid
    //       ? ValidationState.OK
    //       : ValidationState.ERROR,
    //   },
    // ]);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async updateMeasureSqlName(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    sqlName: string
  ) {
    // TODO: validations
    const modifications: Partial<MeasureDefinitionEntity> = {
      sqlName,
      sqlNameIsValid:
        sqlName !== "" ? ValidationState.OK : ValidationState.ERROR,
    };
    // this.dataModelerStateService.dispatch("updateMeasure", [
    //   metricsDefId,
    //   measureId,
    //   modifications,
    // ]);
    // rillRequestContext.actionsChannel.pushMessage("updateMeasure", [
    //   metricsDefId,
    //   measureId,
    //   modifications,
    // ]);
  }
}
