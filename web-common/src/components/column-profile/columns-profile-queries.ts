import type { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

export async function getTableCardinality(
  instanceId: string,
  tableName: string,
  batchedRequest: BatchedRequest
) {
  return batchedRequest.addReq(
    {
      tableCardinalityRequest: {
        instanceId,
        tableName,
        priority: getPriority("table-cardinality"),
      },
    },
    (data) => +data.tableCardinalityResponse.cardinality
  );
}

export async function getNullCounts(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest
) {
  return batchedRequest.addReq(
    {
      columnNullCountRequest: {
        instanceId,
        tableName,
        columnName,
        priority: getPriority("null-count"),
      },
    },
    (data) => +data.columnNullCountResponse.count
  );
}
