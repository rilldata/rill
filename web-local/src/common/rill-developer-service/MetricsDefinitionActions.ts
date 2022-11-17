import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { getName } from "@rilldata/web-local/common/utils/incrementName";
import { CATEGORICALS } from "@rilldata/web-local/lib/duckdb-data-types";
import type { ProfileColumn } from "@rilldata/web-local/lib/types";
import { ActionResponseFactory } from "../data-modeler-service/response/ActionResponseFactory";
import {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { MetricsDefinitionEntity } from "../data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { ExplorerSourceModelDoesntExist } from "../errors/ErrorMessages";
import { getMetricsDefinition } from "../stateInstancesFactory";
import { shallowCopy } from "../utils/shallowCopy";
import { RillDeveloperActions } from "./RillDeveloperActions";
import type { RillRequestContext } from "./RillRequestContext";

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
      getName(
        "dashboard_0",
        this.dataModelerStateService
          .getMetricsDefinitionService()
          .getCurrentState()
          .entities.map((metricsDef) => metricsDef.metricDefLabel)
      )
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
    const measure = measureResp.data as MeasureDefinitionEntity;

    measure.label = "Number of records";
    measure.description = "Number of records in current selection";

    await this.rillDeveloperService.dispatch(
      rillRequestContext,
      "updateMeasure",
      [measure.id, measure]
    );
    rillRequestContext.actionsChannel.pushMessage(
      measureResp.data as Record<string, unknown>
    );
  }
}
