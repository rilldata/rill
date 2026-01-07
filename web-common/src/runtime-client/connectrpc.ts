import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryService } from "../proto/gen/rill/runtime/v1/queries_connect";
import { createQuery, type CreateQueryOptions } from "@tanstack/svelte-query";
import {
  MetricsViewAggregationResponse,
  MetricsViewTimeSeriesResponse,
  type MetricsViewAggregationRequest,
  type MetricsViewTimeSeriesRequest,
} from "../proto/gen/rill/runtime/v1/queries_pb";
import type { PartialMessage } from "@bufbuild/protobuf";
import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import { projectManager } from "../features/project/project-manager";
import { getContext } from "svelte";
import { Timestamp } from "@bufbuild/protobuf";
import { queryClient } from "../lib/svelte-query/globalQueryClient";
import { TimeGrain } from "../proto/gen/rill/runtime/v1/time_grain_pb";
import { V1TimeGrain } from "./gen/index.schemas";

export function useQueryServiceClient(
  { organization, project }: { organization: string; project: string } = {
    organization: getContext("organization"),
    project: getContext("project"),
  },
) {
  return projectManager.getProjectContext({
    organization,
    project,
  });
}

export function createQueryServiceClient(projectData: V1GetProjectResponse) {
  const baseUrl: string =
    projectData.deployment?.runtimeHost || "http://localhost:9009";
  const instanceId: string =
    projectData.deployment?.runtimeInstanceId || "default";
  const token = projectData.jwt;

  const transport = createConnectTransport({
    baseUrl,
    useBinaryFormat: false,
    interceptors: token
      ? [
          (next) => async (req) => {
            req.header.set("Authorization", `Bearer ${token}`);
            return await next(req);
          },
        ]
      : [],
  });

  const client = createClient(QueryService, transport);

  return {
    metricsViewAggregation: (
      params: PartialMessage<MetricsViewAggregationRequest>,
      options?: {
        query?: Partial<
          CreateQueryOptions<
            MetricsViewAggregationResponse,
            Error,
            MetricsViewAggregationResponse
          >
        >;
      },
    ) => {
      const queryOptions = {
        queryKey: [
          "RuntimeService",
          "metricsViewAggregation",
          instanceId,
          params.dimensions,
          params.measures,
        ],
        queryFn: () =>
          client.metricsViewAggregation({
            ...params,
            instanceId,
          }),
        ...options?.query,
        enabled: !!instanceId && (options?.query?.enabled ?? true),
        queryClient: queryClient,
      };

      return {
        create: () => createQuery(queryOptions),
        options: queryOptions,
      };
    },

    metricsViewTimeSeries: (
      params: PartialMessage<MetricsViewTimeSeriesRequest>,
      options?: {
        query?: Partial<
          CreateQueryOptions<
            MetricsViewTimeSeriesResponse,
            Error,
            MetricsViewTimeSeriesResponse
          >
        >;
      },
    ) => {
      const queryOptions = {
        queryKey: [
          "RuntimeService",
          "metricsViewTimeSeries",
          instanceId,
          params.timeStart,
          params.timeEnd,
          params.timeGranularity,
          params.measureNames,
          params.where,
        ],
        queryFn: () =>
          client.metricsViewTimeSeries({
            ...params,
            instanceId,
          }),
        ...options?.query,
        enabled: !!instanceId && (options?.query?.enabled ?? true),
        queryClient: queryClient,
      };

      return {
        create: () => createQuery(queryOptions),
        options: queryOptions,
      };
    },
  };
}

export function toTimestamp(
  dateString: string | Date | undefined,
): Timestamp | undefined {
  if (!dateString) return undefined;
  if (typeof dateString === "string") {
    return Timestamp.fromJson(dateString);
  }
  return Timestamp.fromDate(dateString);
}

export function toConnectRPCTimeGrain(
  timeGrain: V1TimeGrain | undefined,
): TimeGrain {
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_UNSPECIFIED:
      return TimeGrain.UNSPECIFIED;
    case V1TimeGrain.TIME_GRAIN_SECOND:
      return TimeGrain.SECOND;
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return TimeGrain.MINUTE;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return TimeGrain.HOUR;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return TimeGrain.DAY;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return TimeGrain.WEEK;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return TimeGrain.MONTH;
    case V1TimeGrain.TIME_GRAIN_QUARTER:
      return TimeGrain.QUARTER;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return TimeGrain.YEAR;
    default:
      return TimeGrain.UNSPECIFIED;
  }
}
