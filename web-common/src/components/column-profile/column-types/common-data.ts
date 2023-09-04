import type { ColumnsProfileDataUpdate } from "@rilldata/web-common/components/column-profile/columns-profile-data";
import type { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

export async function loadTableCardinality(
  instanceId: string,
  tableName: string,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const tableCardinality = await batchedRequest.add(
    {
      tableCardinalityRequest: {
        instanceId,
        tableName,
        priority: getPriority("table-cardinality"),
      },
    },
    (data) => Number(data.tableCardinalityResponse.cardinality ?? 0)
  );
  update((state) => {
    // TODO: why is this called table rows where as it sets cardinality
    state.tableRows = tableCardinality;
    state.isFetching = false;
    return state;
  });
}

export async function loadColumnsNullCount(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const nullCount = await batchedRequest.add(
    {
      columnNullCountRequest: {
        instanceId,
        tableName,
        columnName,
        priority: getPriority("null-count"),
      },
    },
    (data) => data.columnNullCountResponse.count
  );
  update((state) => {
    if (!state.profiles[columnName]) return;
    state.profiles[columnName].nullCount = nullCount;
    return state;
  });
}

export async function loadColumnCardinality(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const columnCardinality = await batchedRequest.add(
    {
      columnCardinalityRequest: {
        instanceId,
        tableName,
        columnName,
        priority: getPriority("column-cardinality"),
      },
    },
    (data) => data.columnCardinalityResponse.categoricalSummary.cardinality
  );
  update((state) => {
    if (!state.profiles[columnName]) return;
    state.profiles[columnName].cardinality = columnCardinality;
    return state;
  });
}

export async function loadColumnTopK(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const topK = await batchedRequest.add(
    {
      columnTopKRequest: {
        instanceId,
        tableName,
        columnName,
        agg: "count(*)",
        k: 75,
        priority: getPriority("topk"),
      },
    },
    (data) => data.columnTopKResponse.categoricalSummary.topK.entries
  );
  update((state) => {
    if (!state.profiles[columnName]) return;
    state.profiles[columnName].topK = topK;
    return state;
  });
}
