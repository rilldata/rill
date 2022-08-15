import { RillDeveloperActions } from "$common/rill-developer-service/RillDeveloperActions";
import type { MetricsDefinitionContext } from "$common/rill-developer-service/MetricsDefinitionActions";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import { ActionResponseFactory } from "$common/data-modeler-service/response/ActionResponseFactory";
import type { TimeSeriesRollup } from "$common/database-service/DatabaseTimeSeriesActions";
import { DatabaseActionQueuePriority } from "$common/priority-action-queue/DatabaseActionQueuePriority";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { TimeSeriesValue } from "$lib/redux-store/timeseries/timeseries-slice";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExplorerActions";

export interface RuntimeMetricsMetaResponse {
  name: string;
  timeDimension: {
    name: string;
    timeRange: TimeSeriesTimeRange;
  };
  dimensions: Array<DimensionDefinitionEntity>;
  measures: Array<MeasureDefinitionEntity>;
}

export interface RuntimeRequestTimeRange {
  start: string;
  end: string;
  granularity: string;
}
export interface RuntimeDimensionValue {
  name: string;
  values: Array<unknown>;
}
export type RuntimeDimensionValues = Array<RuntimeDimensionValue>;
export interface RuntimeRequestFilter {
  include: RuntimeDimensionValues;
  exclude: RuntimeDimensionValues;
}

export interface RuntimeTimeSeriesRequest {
  measures: Array<string>;
  time: RuntimeRequestTimeRange;
  filter: RuntimeRequestFilter;
}
export interface RuntimeTimeSeriesResponse {
  meta: Array<{ name: string; type: string }>;
  // data: Array<{ time: string } & Record<string, number>>;
  data: Array<TimeSeriesValue>;
}

export interface RuntimeTopListRequest {
  measures: Array<string>;
  time: Pick<RuntimeRequestTimeRange, "start" | "end">;
  limit: number;
  offset: number;
  sort: Array<{ name: string; direction: "desc" | "asc" }>;
  filter: RuntimeRequestFilter;
}
export interface RuntimeTopListResponse {
  meta: Array<{ name: string; type: string }>;
  // data: Array<Record<string, number | string>>;
  data: Array<{ label: string; value: number }>;
}

export interface RuntimeBigNumberRequest {
  measures: Array<string>;
  time: Pick<RuntimeRequestTimeRange, "start" | "end">;
  filter: RuntimeRequestFilter;
}
export interface RuntimeBigNumberResponse {
  meta: Array<{ name: string; type: string }>;
  data: Record<string, number>;
}

function convertToActiveValues(filters: RuntimeRequestFilter): ActiveValues {
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

/**
 * Actions that get info for metrics explore.
 * Based on rill runtime specs.
 */
export class RillRuntimeMetricsActions extends RillDeveloperActions {
  @RillDeveloperActions.MetricsDefinitionAction()
  public async getRuntimeMetricsMeta(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string
  ) {
    // TODO: validation
    const timeRange: TimeSeriesTimeRange = (
      await this.rillDeveloperService.dispatch(
        rillRequestContext,
        "getTimeRange",
        [metricsDefId]
      )
    ).data;
    const meta: RuntimeMetricsMetaResponse = {
      name: rillRequestContext.record.metricDefLabel,
      timeDimension: {
        name: rillRequestContext.record.timeDimension,
        timeRange,
      },
      measures: this.dataModelerStateService
        .getMeasureDefinitionService()
        .getManyByField("metricsDefId", metricsDefId),
      dimensions: this.dataModelerStateService
        .getDimensionDefinitionService()
        .getManyByField("metricsDefId", metricsDefId),
    };
    return ActionResponseFactory.getRawResponse(meta);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getRuntimeTimeSeries(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    request: RuntimeTimeSeriesRequest
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
          measures: request.measures.map((measureId) =>
            this.dataModelerStateService
              .getMeasureDefinitionService()
              .getById(measureId)
          ),
          filters: convertToActiveValues(request.filter),
          timeRange: request.time,
        },
      ]
    );
    const response: RuntimeTimeSeriesResponse = {
      meta: [], // TODO
      data: timeSeries.rollup.results,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getRuntimeTopList(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    dimensionId: string,
    request: RuntimeTopListRequest
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
        convertToActiveValues(request.filter),
        rillRequestContext.record.timeDimension,
        request.time,
      ]
    );
    const response: RuntimeTopListResponse = {
      meta: [], // TODO
      data,
    };
    return ActionResponseFactory.getRawResponse(response);
  }

  @RillDeveloperActions.MetricsDefinitionAction()
  public async getRuntimeBigNumber(
    rillRequestContext: MetricsDefinitionContext,
    metricsDefId: string,
    request: RuntimeBigNumberRequest
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
          request.measures.map((measureId) =>
            this.dataModelerStateService
              .getMeasureDefinitionService()
              .getById(measureId)
          ),
          convertToActiveValues(request.filter),
          rillRequestContext.record.timeDimension,
          request.time,
        ]
      );
    const response: RuntimeBigNumberResponse = {
      meta: [], // TODO
      data: bigNumberResponse.bigNumbers,
    };
    return ActionResponseFactory.getRawResponse(response);
  }
}
