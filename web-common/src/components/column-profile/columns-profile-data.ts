import {
  loadColumnCardinality,
  loadColumnsNullCount,
  loadColumnTopK,
  loadTableCardinality,
} from "@rilldata/web-common/components/column-profile/column-types/common-data";
import {
  loadColumnHistogram,
  loadDescriptiveStatistics,
} from "@rilldata/web-common/components/column-profile/column-types/numeric-profile-data";
import { loadTimeSeries } from "@rilldata/web-common/components/column-profile/column-types/timestamp-profile-data";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import {
  CATEGORICALS,
  INTERVALS,
  isFloat,
  isNested,
  NUMERICS,
  TIMESTAMPS,
} from "@rilldata/web-common/lib/duckdb-data-types";
import type {
  NumericHistogramBinsBin,
  NumericOutliersOutlier,
  TopKEntry,
  V1NumericStatistics,
  V1TableColumnsResponse,
  V1TimeGrain,
  V1TimeSeriesValue,
} from "@rilldata/web-common/runtime-client";
import { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import type { QueryObserverResult } from "@tanstack/query-core";
import { getContext, setContext } from "svelte";
import { Updater, writable } from "svelte/store";
import type { Readable } from "svelte/store";

export type ColumnProfileData = {
  name: string;
  type: string;

  isFetching: boolean;
  nullCount?: number;
  cardinality?: number;

  topK?: Array<TopKEntry>;

  // numeric profile
  rugHistogram?: Array<NumericOutliersOutlier>;
  histogram?: Array<NumericHistogramBinsBin>;
  descriptiveStatistics?: V1NumericStatistics;

  // timestamp profile
  estimatedRollupInterval?: V1TimeGrain;
  smallestTimeGrain?: V1TimeGrain;
  timeSeriesData?: Array<V1TimeSeriesValue>;
  timeSeriesSpark?: Array<V1TimeSeriesValue>;
};
export type ColumnsProfileData = {
  isFetching: boolean;
  tableRows: number;
  columnNames: Array<string>;
  profiles: Record<string, ColumnProfileData>;
};
export type ColumnsProfileDataMethods = {
  load: (
    instanceId: string,
    tableName: string,
    profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>
  ) => Promise<void>;
};
export type ColumnsProfileDataStore = Readable<ColumnsProfileData> &
  ColumnsProfileDataMethods;

export function setColumnsProfileStore(store: ColumnsProfileDataStore) {
  setContext("COLUMNS_PROFILE", store);
}

export function getColumnsProfileStore() {
  return getContext<ColumnsProfileDataStore>("COLUMNS_PROFILE");
}

export function createColumnsProfileData(): ColumnsProfileDataStore {
  const { update, subscribe } = writable<ColumnsProfileData>({
    isFetching: true,
    tableRows: 0,
    columnNames: [],
    profiles: {},
  });

  let batchedRequest: BatchedRequest;

  return {
    subscribe,
    load: async (
      instanceId: string,
      tableName: string,
      profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>
    ) => {
      batchedRequest?.cancel();

      resetState(profileColumnResponse, update);

      batchedRequest = new BatchedRequest();
      loadTableCardinality(instanceId, tableName, batchedRequest, update);

      for (const column of profileColumnResponse.data.profileColumns) {
        const columnName = column.name;
        const columnPromises = new Array<Promise<any>>();
        columnPromises.push(
          loadColumnsNullCount(
            instanceId,
            tableName,
            columnName,
            batchedRequest,
            update
          ),
          loadColumnCardinality(
            instanceId,
            tableName,
            columnName,
            batchedRequest,
            update
          )
        );

        let type = column.type;
        if (!type) continue;
        if (type.includes("DECIMAL")) type = "DECIMAL";

        if (CATEGORICALS.has(type)) {
          columnPromises.push(
            loadColumnTopK(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              update
            )
          );
        } else if (NUMERICS.has(type) || INTERVALS.has(type)) {
          columnPromises.push(
            loadColumnHistogram(
              instanceId,
              tableName,
              columnName,
              isFloat(type),
              batchedRequest,
              update
            ),
            loadDescriptiveStatistics(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              update
            )
          );
        } else if (TIMESTAMPS.has(type)) {
          columnPromises.push(
            loadTimeSeries(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              update
            )
          );
        } else if (isNested(type)) {
          columnPromises.push(
            loadColumnTopK(
              instanceId,
              tableName,
              columnName,
              batchedRequest,
              update
            )
          );
        }

        Promise.all(columnPromises).then(() => {
          update((state) => {
            if (!state.profiles[columnName]) return;
            state.profiles[columnName].isFetching = false;
            return state;
          });
        });
      }

      return batchedRequest.send(instanceId);
    },
  };
}

export type ColumnsProfileDataUpdate = (
  this: void,
  updater: Updater<ColumnsProfileData>
) => void;

export function resetState(
  profileColumnResponse: QueryObserverResult<V1TableColumnsResponse>,
  update: ColumnsProfileDataUpdate
) {
  const columnsMap = getMapFromArray(
    profileColumnResponse.data.profileColumns,
    (entity) => entity.name
  );

  update((state) => {
    state.isFetching = true;

    // remove older columns
    for (const oldColumnName in state.profiles) {
      if (!columnsMap.has(oldColumnName)) {
        delete state.profiles[oldColumnName];
      }
    }

    const columnNames = new Array<string>();

    // mark everything as fetching
    for (const column of profileColumnResponse.data.profileColumns) {
      if (!(column.name in state.profiles)) {
        state.profiles[column.name] = {
          name: column.name,
          type: column.type,
          isFetching: true,
        };
      } else {
        state.profiles[column.name].isFetching = true;
      }
      columnNames.push(column.name);
    }

    state.columnNames = columnNames;

    return state;
  });
}
