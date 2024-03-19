import { error } from "@sveltejs/kit";
import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client/index.js";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.js";
import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { CATALOG_ENTRY_NOT_FOUND } from "@rilldata/web-local/lib/errors/messages.js";
import { createQueryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient.js";

export function load({ params }) {
  //   if (params.name === "hello-world") {
  const filePath = getFilePathFromNameAndType(
    params.name,
    EntityType.MetricsDefinition,
  );

  createRuntimeServiceGetFile("default", filePath, {
    query: {
      queryClient: createQueryClient(),
      onError: (err) => {
        if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
          error(404, "Dashboard not found");
        }

        throw error(err.response?.status || 500, err.message);
      },
    },
  })

 
}
