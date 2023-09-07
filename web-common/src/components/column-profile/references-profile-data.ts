import type { ColumnsProfileDataUpdate } from "@rilldata/web-common/components/column-profile/columns-profile-data";
import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
import type { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

export async function loadReferencesProfile(
  instanceId: string,
  references: Array<V1CatalogEntry>,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  return Promise.all(
    references.map((reference, i) =>
      loadSingleReferenceProfile(
        instanceId,
        reference,
        i,
        batchedRequest,
        update
      )
    )
  );
}

async function loadSingleReferenceProfile(
  instanceId: string,
  reference: V1CatalogEntry,
  index: number,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const cardinalityPromise = batchedRequest.add(
    {
      tableCardinalityRequest: {
        instanceId,
        tableName: reference.name,
        priority: getPriority("table-cardinality"),
      },
    },
    (data) => +data?.tableCardinalityResponse?.cardinality
  );
  const columnsPromise = batchedRequest.add(
    {
      tableColumnsRequest: {
        instanceId,
        tableName: reference.name,
        priority: getPriority("columns-profile"),
      },
    },
    (data) => data?.tableColumnsResponse.profileColumns
  );

  const [cardinality, columns] = await Promise.all([
    cardinalityPromise,
    columnsPromise,
  ]);
  update((state) => {
    state.references[index] = {
      cardinality,
      columns,
    };
    return state;
  });
}
