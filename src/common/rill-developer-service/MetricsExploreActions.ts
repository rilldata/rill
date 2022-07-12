import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  BasicMeasureDefinition,
  MeasureDefinitionEntity,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { TimeSeriesRollup } from "$common/database-service/DatabaseTimeSeriesActions";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";

export class MetricsExploreActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async generateTimeSeries(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    {
      measures,
      filters,
      pixels,
      rollupInterval,
      isolated,
    }: {
      measures: Array<BasicMeasureDefinition>;
      filters: ActiveValues;
      pixels: number;
      rollupInterval: RollupInterval;
      isolated?: boolean;
    }
  ) {
    if (
      !rillRequestContext.record?.sourceModelId ||
      !rillRequestContext.record?.timeDimension
    )
      return;

    if (isolated) {
      await Promise.all(
        measures.map((measure) =>
          this.generateTimeSeriesForMeasures(rillRequestContext, measure.id, {
            measures: [measure],
            filters,
            pixels,
            rollupInterval,
          })
        )
      );
    } else {
      await this.generateTimeSeriesForMeasures(
        rillRequestContext,
        metricsDefId,
        { measures, filters, pixels, rollupInterval }
      );
    }
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getLeaderboardValues(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    measureId: string,
    filters: ActiveValues
  ) {
    if (!rillRequestContext.record?.sourceModelId) return;
    const measure = this.dataModelerStateService
      .getMeasureDefinitionService()
      .getById(measureId);
    const dimensions = this.dataModelerStateService
      .getDimensionDefinitionService()
      .getCurrentState()
      .entities.filter((dimension) => dimension.metricsDefId === metricsDefId);
    await Promise.all(
      dimensions.map((dimension) =>
        this.getLeaderboardValuesForDimension(
          rillRequestContext,
          measure,
          dimension,
          filters
        )
      )
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getBigNumber(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    {
      filters,
      measures,
      isolated,
    }: {
      filters: ActiveValues;
      measures: Array<BasicMeasureDefinition>;
      isolated?: boolean;
    }
  ) {
    if (!rillRequestContext.record?.sourceModelId) return;
    if (isolated) {
      await Promise.all(
        measures.map((measure) =>
          this.generateBigNumberForMeasures(rillRequestContext, measure.id, {
            filters,
            measures: [measure],
          })
        )
      );
    } else {
      await this.generateBigNumberForMeasures(
        rillRequestContext,
        metricsDefId,
        { filters, measures }
      );
    }
  }

  private async generateTimeSeriesForMeasures(
    rillRequestContext: MetricsDefinitionContext,
    id: string,
    {
      measures,
      filters,
      pixels,
      rollupInterval,
    }: {
      measures: Array<BasicMeasureDefinition>;
      filters: ActiveValues;
      pixels: number;
      rollupInterval: RollupInterval;
    }
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const timeSeries: TimeSeriesRollup = await this.databaseActionQueue.enqueue(
      {
        id,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "generateTimeSeries",
      [
        {
          tableName: model.tableName,
          timestampColumn: rillRequestContext.record.timeDimension,
          measures,
          filters,
          pixels,
          rollupInterval,
        },
      ]
    );
    timeSeries.rollup.id = id;
    rillRequestContext.actionsChannel.pushMessage(timeSeries.rollup as any);
  }

  private async getLeaderboardValuesForDimension(
    rillRequestContext: MetricsDefinitionContext,
    measure: MeasureDefinitionEntity,
    dimension: DimensionDefinitionEntity,
    filters: ActiveValues
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    rillRequestContext.actionsChannel.pushMessage({
      dimensionId: dimension.id,
      values: await this.databaseActionQueue.enqueue(
        {
          id: rillRequestContext.id,
          priority: DatabaseActionQueuePriority.ActiveModel,
        },
        "getLeaderboardValues",
        [
          model.tableName,
          dimension.dimensionColumn,
          measure.expression,
          filters,
        ]
      ),
    });
  }

  private async generateBigNumberForMeasures(
    rillRequestContext: MetricsDefinitionContext,
    id: string,
    {
      measures,
      filters,
    }: {
      measures: Array<BasicMeasureDefinition>;
      filters: ActiveValues;
    }
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const bigNumbers = await this.databaseActionQueue.enqueue(
      {
        id: rillRequestContext.id,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "getBigNumber",
      [model.tableName, measures, filters]
    );
    bigNumbers.id = id;
    rillRequestContext.actionsChannel.pushMessage(bigNumbers);
  }
}
