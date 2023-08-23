import type { ColumnsProfileDataUpdate } from "@rilldata/web-common/components/column-profile/columns-profile-data";
import { convertTimestampPreview } from "@rilldata/web-common/lib/convertTimestampPreview";
import type { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

export async function loadTimeSeries(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const timeSeriesPromise = batchedRequest.addReq(
    {
      columnTimeSeriesRequest: {
        instanceId,
        tableName,
        timestampColumnName: columnName,
        measures: [{ expression: "count(*)" }],
        pixels: 92,
        priority: getPriority("timeseries"),
      },
    },
    (data) => data.columnTimeSeriesResponse.rollup
  );
  const rollupIntervalPromise = batchedRequest.addReq(
    {
      columnRollupIntervalRequest: {
        instanceId,
        tableName,
        columnName,
      },
    },
    (data) => data.columnRollupIntervalResponse.interval
  );
  const smallestTimeGrainPromise = batchedRequest.addReq(
    {
      columnTimeGrainRequest: {
        instanceId,
        tableName,
        columnName,
      },
    },
    (data) => data.columnTimeGrainResponse.timeGrain
  );

  const [timeSeries, rollupInterval, smallestTimeGrain] = await Promise.all([
    timeSeriesPromise,
    rollupIntervalPromise,
    smallestTimeGrainPromise,
  ]);

  update((state) => {
    if (!state.profiles[columnName]) return;
    state.profiles[columnName].estimatedRollupInterval = rollupInterval;
    state.profiles[columnName].smallestTimeGrain = smallestTimeGrain;
    state.profiles[columnName].timeSeriesData = convertTimestampPreview(
      timeSeries.results?.map((di) => {
        const next = { ...di, count: di.records.count };
        if (next.count == null || !isFinite(next.count)) {
          next.count = 0;
        }
        return next;
      }) || []
    );
    state.profiles[columnName].timeSeriesSpark = convertTimestampPreview(
      timeSeries.spark?.map((di) => {
        const next = { ...di, count: di.records.count };
        if (next.count == null || !isFinite(next.count)) {
          next.count = 0;
        }
        return next;
      }) || []
    );
    return state;
  });
}
