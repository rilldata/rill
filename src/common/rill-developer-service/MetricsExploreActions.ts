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
import type {
  TimeSeriesRollup,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";

export class MetricsExploreActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async getTimeRange(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    if (
      !rillRequestContext.record?.sourceModelId ||
      !rillRequestContext.record?.timeDimension
    )
      return;
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const rollupInterval: RollupInterval =
      await this.databaseActionQueue.enqueue(
        {
          id: metricsDefId,
          priority: DatabaseActionQueuePriority.ActiveModel,
        },
        "estimateIdealRollupInterval",
        [model.tableName, rillRequestContext.record.timeDimension]
      );
    return ActionResponseFactory.getSuccessResponse("", {
      interval: rollupInterval.rollupInterval,
      start: rollupInterval.minValue,
      end: rollupInterval.maxValue,
    } as TimeSeriesTimeRange);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async generateTimeSeries(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    {
      measures,
      filters,
      pixels,
      timeRange,
      isolated,
    }: {
      measures: Array<BasicMeasureDefinition>;
      filters: ActiveValues;
      pixels: number;
      timeRange?: TimeSeriesTimeRange;
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
            timeRange,
          })
        )
      );
    } else {
      await this.generateTimeSeriesForMeasures(
        rillRequestContext,
        metricsDefId,
        { measures, filters, pixels, timeRange }
      );
    }
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getLeaderboardValues(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    {
      measureId,
      filters,
      timeRange,
    }: {
      measureId: string;
      filters: ActiveValues;
      timeRange?: TimeSeriesTimeRange;
    }
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
          filters,
          timeRange
        )
      )
    );
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getBigNumber(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    {
      measures,
      filters,
      timeRange,
      isolated,
    }: {
      measures: Array<BasicMeasureDefinition>;
      filters: ActiveValues;
      timeRange?: TimeSeriesTimeRange;
      isolated?: boolean;
    }
  ) {
    if (!rillRequestContext.record?.sourceModelId) return;
    if (isolated) {
      await Promise.all(
        measures.map((measure) =>
          this.generateBigNumberForMeasures(
            rillRequestContext,
            measure.id,
            [measure],
            filters,
            timeRange
          )
        )
      );
    } else {
      await this.generateBigNumberForMeasures(
        rillRequestContext,
        metricsDefId,
        measures,
        filters,
        timeRange
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
      timeRange,
    }: {
      measures: Array<BasicMeasureDefinition>;
      filters: ActiveValues;
      pixels: number;
      timeRange?: TimeSeriesTimeRange;
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
          timeRange,
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
    filters: ActiveValues,
    timeRange: TimeSeriesTimeRange
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    rillRequestContext.actionsChannel.pushMessage({
      dimensionId: dimension.id,
      dimensionName: dimension.dimensionColumn,
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
          rillRequestContext.record.timeDimension,
          timeRange,
        ]
      ),
    });
  }

  private async generateBigNumberForMeasures(
    rillRequestContext: MetricsDefinitionContext,
    id: string,
    measures: Array<BasicMeasureDefinition>,
    filters: ActiveValues,
    timeRange: TimeSeriesTimeRange
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
      [
        model.tableName,
        measures,
        filters,
        rillRequestContext.record.timeDimension,
        timeRange,
      ]
    );
    bigNumbers.id = id;
    rillRequestContext.actionsChannel.pushMessage(bigNumbers);
  }
}
