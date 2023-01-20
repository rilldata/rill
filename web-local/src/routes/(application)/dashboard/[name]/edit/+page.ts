import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import {
  runtimeServiceGetCatalogEntry,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { error } from "@sveltejs/kit";
import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const config = await runtimeServiceGetConfig();
  if (config.readonly) {
    throw error(404, "Page not found");
  }

  try {
    await runtimeServiceGetFile(
      config.instance_id,
      getFilePathFromNameAndType(params.name, EntityType.MetricsDefinition)
    );
  } catch (err) {
    if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
      throw error(404, "Dashboard not found");
    }

    throw error(err.response?.status || 500, err.message);
  }

  try {
    await runtimeServiceGetCatalogEntry(config.instance_id, params.name);

    return {
      metricsDefName: params.name,
    };
  } catch (err) {
    // If the catalog entry doesn't exist, the dashboard config is invalid
    // The component should render the specific error
    if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
      return {
        metricsDefName: params.name,
        error: err.message,
      };
    }

    // Throw all other errors
    throw error(err.response?.status || 500, err.message);
  }
}
