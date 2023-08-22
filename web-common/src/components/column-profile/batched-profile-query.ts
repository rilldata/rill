import {
  createBatchedColumnTimeGrainQuery,
  createBatchedColumnCardinality,
  createBatchedColumnDescriptiveStatisticsQuery,
  createBatchedColumnNullCount,
  createBatchedColumnNumericHistogramQuery,
  createBatchedColumnRugHistogramQuery,
  createBatchedColumnTimeSeriesQuery,
  createBatchedColumnTopKQuery,
  createBatchedServiceColumnRollupIntervalQuery,
} from "@rilldata/web-common/components/column-profile/batched-queries";
import {
  CATEGORICALS,
  INTERVALS,
  isFloat,
  isNested,
  NUMERICS,
  TIMESTAMPS,
} from "@rilldata/web-common/lib/duckdb-data-types";
import { QueryServiceColumnNumericHistogramHistogramMethod } from "@rilldata/web-common/runtime-client";
import type { V1TableColumnsResponse } from "@rilldata/web-common/runtime-client";
import { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { waitUntil } from "@rilldata/web-local/lib/util/waitUtils";
import type { QueryObserverResult } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, readable } from "svelte/store";

export function batchedProfileQuery(
  instanceId: string,
  tableName: string,
  profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>,
  done: () => void
) {
  if (!profileColumnResponse?.data || profileColumnResponse.isFetching)
    return readable(false);

  const batchedRequest = new BatchedRequest();
  const stores = new Array<CreateQueryResult>();
  for (const column of profileColumnResponse.data.profileColumns) {
    stores.push(
      createBatchedColumnNullCount(
        instanceId,
        tableName,
        column.name,
        batchedRequest
      )
    );
    stores.push(
      createBatchedColumnCardinality(
        instanceId,
        tableName,
        column.name,
        batchedRequest
      )
    );

    let type = column.type;
    if (!type) continue;
    if (type.includes("DECIMAL")) type = "DECIMAL";
    if (CATEGORICALS.has(type)) {
      stores.push(
        ...addCategoricalProfiles(
          instanceId,
          tableName,
          column.name,
          batchedRequest
        )
      );
    } else if (NUMERICS.has(type) || INTERVALS.has(type)) {
      stores.push(
        ...addNumericProfiles(
          instanceId,
          tableName,
          column.name,
          type,
          batchedRequest
        )
      );
    } else if (TIMESTAMPS.has(type)) {
      stores.push(
        ...addTimestampProfiles(
          instanceId,
          tableName,
          column.name,
          batchedRequest
        )
      );
    } else if (isNested(type)) {
      stores.push(
        ...addNestedProfiles(instanceId, tableName, column.name, batchedRequest)
      );
    }
  }
  waitUntil(() => batchedRequest.ready)
    .then(() => batchedRequest.send(instanceId))
    .then(done);

  return derived(stores, () => true);
}

function addCategoricalProfiles(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return [
    createBatchedColumnTopKQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
  ];
}

function addNumericProfiles(
  instanceId: string,
  tableName: string,
  columnName: string,
  type: string,
  batchedRequest: BatchedRequest
) {
  return [
    createBatchedColumnTopKQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),

    createBatchedColumnNumericHistogramQuery(
      instanceId,
      tableName,
      columnName,
      QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_DIAGNOSTIC,
      batchedRequest
    ),
    ...(isFloat(type)
      ? [
          createBatchedColumnNumericHistogramQuery(
            instanceId,
            tableName,
            columnName,
            QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_FD,
            batchedRequest
          ),
        ]
      : []),

    createBatchedColumnRugHistogramQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
    createBatchedColumnDescriptiveStatisticsQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
  ];
}

function addTimestampProfiles(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return [
    createBatchedColumnTimeSeriesQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
    createBatchedServiceColumnRollupIntervalQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
    createBatchedColumnTimeGrainQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
  ];
}

function addNestedProfiles(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return [
    createBatchedColumnTopKQuery(
      instanceId,
      tableName,
      columnName,
      batchedRequest
    ),
  ];
}
