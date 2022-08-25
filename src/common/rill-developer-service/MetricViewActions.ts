import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { getFallbackMeasureName } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExplorerActions";
import type {
  TimeSeriesRollup,
  TimeSeriesTimeRange,
  TimeSeriesValue,
} from "$common/database-service/DatabaseTimeSeriesActions";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import { RillRequestContext } from "$common/rill-developer-service/RillRequestContext";
import { getMapFromArray } from "$common/utils/getMapFromArray";
import type { ActiveValues } from "$lib/application-state-stores/explorer-stores";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";

export interface MetricViewMetaResponse {
  name: string;
  timeDimension: {
    name: string;
    timeRange: TimeSeriesTimeRange;
  };
  dimensions: Array<DimensionDefinitionEntity>;
  measures: Array<MeasureDefinitionEntity>;
}

export interface MetricViewRequestTimeRange {
  start: string;
  end: string;
  granularity: string;
}
export interface MetricViewDimensionValue {
  name: string;
  values: Array<unknown>;
}
export type MetricViewDimensionValues = Array<MetricViewDimensionValue>;
export interface MetricViewRequestFilter {
  include: MetricViewDimensionValues;
  exclude: MetricViewDimensionValues;
}

export interface MetricViewTimeSeriesRequest {
  measures: Array<string>;
  time: MetricViewRequestTimeRange;
  filter?: MetricViewRequestFilter;
}
export interface MetricViewTimeSeriesResponse {
  meta: Array<{ name: string; type: string }>;
  // data: Array<{ time: string } & Record<string, number>>;
  data: Array<TimeSeriesValue>;
}

export interface MetricViewTopListRequest {
  measures: Array<string>;
  time: Pick<MetricViewRequestTimeRange, "start" | "end">;
  limit: number;
  offset: number;
  sort: Array<{ name: string; direction: "desc" | "asc" }>;
  filter?: MetricViewRequestFilter;
}
export interface MetricViewTopListResponse {
  meta: Array<{ name: string; type: string }>;
  // data: Array<Record<string, number | string>>;
  data: Array<{ label: string; value: number }>;
}

export interface MetricViewTotalsRequest {
  measures: Array<string>;
  time: Pick<MetricViewRequestTimeRange, "start" | "end">;
  filter?: MetricViewRequestFilter;
}
export interface MetricViewTotalsResponse {
  meta: Array<{ name: string; type: string }>;
  data: Record<string, number>;
}

function convertToActiveValues(filters: MetricViewRequestFilter): ActiveValues {
  if (!filters) return {};
  const activeValues: ActiveValues = {};
  filters.include.forEach((value) => {
    activeValues[value.name] = value.values.map((val) => [val, true]);
  });
  filters.exclude.forEach((value) => {
    activeValues[value.name] ??= [];
    activeValues[value.name].push(
      ...(value.values.map((val) => [val, false]) as Array<[unknown, boolean]>)
    );
  });
  return activeValues;
}

function mapDimensionIdToName(
  filters: MetricViewRequestFilter,
  dimensions: Array<DimensionDefinitionEntity>
): MetricViewRequestFilter {
  if (!filters) return undefined;
  const dimensionsIdMap = getMapFromArray(dimensions, (d) => d.id);
  filters.include.forEach((value) => {
    value.name = dimensionsIdMap.get(value.name).dimensionColumn;
  });
  filters.exclude.forEach((value) => {
    value.name = dimensionsIdMap.get(value.name).dimensionColumn;
  });
  return filters;
}

/**
 * Actions that get info for metrics explore.
 * Based on rill runtime specs.
 */
export class MetricViewActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewMeta(
    rillRequestContext: MetricsDefinitionContext,
    _: string
  ) {
    // TODO: validation
    const meta: MetricViewMetaResponse = {
      name: rillRequestContext.record.metricDefLabel,
      timeDimension: {
        name: rillRequestContext.record.timeDimension,
        timeRange: await this.getTimeRange(rillRequestContext.record),
      },
      measures: await this.getValidMeasures(rillRequestContext.record),
      dimensions: this.getValidDimensions(rillRequestContext.record),
    };
    return ActionResponseFactory.getRawResponse(meta);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewTimeSeries(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    request: MetricViewTimeSeriesRequest
  ) {
    // TODO: validation
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const timeSeries: TimeSeriesRollup = await this.databaseActionQueue.enqueue(
      {
        id: metricsDefId,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "generateTimeSeries",
      [
        {
          tableName: model.tableName,
          timestampColumn: rillRequestContext.record.timeDimension,
          measures: request.measures.map((measureId) => ({
            ...this.dataModelerStateService
              .getMeasureDefinitionService()
              .getById(measureId),
          })),
          filters: convertToActiveValues(
            mapDimensionIdToName(
              request.filter,
              this.dataModelerStateService
                .getDimensionDefinitionService()
                .getManyByField("metricsDefId", metricsDefId)
            )
          ),
          timeRange: {
            ...request.time,
            interval: request.time.granularity,
          },
        },
      ]
    );
    const response: MetricViewTimeSeriesResponse = {
      meta: [], // TODO
      data: timeSeries.rollup.results,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewTopList(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    dimensionId: string,
    request: MetricViewTopListRequest
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const measure = this.dataModelerStateService
      .getMeasureDefinitionService()
      .getById(request.measures[0]);
    const dimension = this.dataModelerStateService
      .getDimensionDefinitionService()
      .getById(dimensionId);
    const data = await this.databaseActionQueue.enqueue(
      {
        id: rillRequestContext.id,
        priority: DatabaseActionQueuePriority.ActiveModel,
      },
      "getLeaderboardValues",
      [
        model.tableName,
        dimension.dimensionColumn,
        measure.expression,
        mapDimensionIdToName(
          request.filter,
          this.dataModelerStateService
            .getDimensionDefinitionService()
            .getManyByField("metricsDefId", metricsDefId)
        ),
        rillRequestContext.record.timeDimension,
        request.time,
      ]
    );
    const response: MetricViewTopListResponse = {
      meta: [], // TODO
      data,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getMetricViewTotals(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    request: MetricViewTotalsRequest
  ) {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(rillRequestContext.record.sourceModelId);
    const bigNumberResponse: BigNumberResponse =
      await this.databaseActionQueue.enqueue(
        {
          id: rillRequestContext.id,
          priority: DatabaseActionQueuePriority.ActiveModel,
        },
        "getBigNumber",
        [
          model.tableName,
          request.measures.map((measureId) => ({
            ...this.dataModelerStateService
              .getMeasureDefinitionService()
              .getById(measureId),
          })),
          mapDimensionIdToName(
            request.filter,
            this.dataModelerStateService
              .getDimensionDefinitionService()
              .getManyByField("metricsDefId", metricsDefId)
          ),
          rillRequestContext.record.timeDimension,
          request.time,
        ]
      );
    const response: MetricViewTotalsResponse = {
      meta: [], // TODO
      data: bigNumberResponse.bigNumbers,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  private async getTimeRange(
    metricsDef: MetricsDefinitionEntity
  ): Promise<TimeSeriesTimeRange> {
    const model = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Persistent)
      .getById(metricsDef.sourceModelId);
    const rollupInterval: RollupInterval =
      await this.databaseActionQueue.enqueue(
        {
          id: metricsDef.id,
          priority: DatabaseActionQueuePriority.ActiveModel,
        },
        "estimateIdealRollupInterval",
        [model.tableName, metricsDef.timeDimension]
      );
    return {
      interval: rollupInterval.rollupInterval,
      start: rollupInterval.minValue,
      end: rollupInterval.maxValue,
    } as TimeSeriesTimeRange;
  }

  private getValidDimensions(metricsDef: MetricsDefinitionEntity) {
    const derivedModel = this.dataModelerStateService
      .getEntityStateService(EntityType.Model, StateType.Derived)
      .getById(metricsDef.sourceModelId);
    if (!derivedModel) {
      return [];
    }

    const columnMap = getMapFromArray(
      derivedModel.profile,
      (column) => column.name
    );

    return this.dataModelerStateService
      .getDimensionDefinitionService()
      .getManyByField("metricsDefId", metricsDef.id)
      .filter((dimension) => columnMap.has(dimension.dimensionColumn));
  }

  private async getValidMeasures(metricsDef: MetricsDefinitionEntity) {
    const measures = this.dataModelerStateService
      .getMeasureDefinitionService()
      .getManyByField("metricsDefId", metricsDef.id);
    return (
      await Promise.all(
        measures.map(async (measure, index) => {
          const measureValidation = await this.rillDeveloperService.dispatch(
            RillRequestContext.getNewContext(),
            "validateMeasureExpression",
            [metricsDef.id, measure.expression]
          );
          return {
            ...measure,
            ...(measureValidation.data as MeasureDefinitionEntity),
            sqlName: getFallbackMeasureName(index, measure.sqlName),
          };
        })
      )
    ).filter((measure) => measure.expressionIsValid === ValidationState.OK);
  }
}
