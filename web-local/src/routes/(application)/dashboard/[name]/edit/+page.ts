import {
  runtimeServiceGetCatalogEntry,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { EntityType } from "@rilldata/web-local/lib/temp/entity";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import { error } from "@sveltejs/kit";
import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await runtimeServiceGetConfig();

  try {
    await runtimeServiceGetFile(
      localConfig.instance_id,
      getFilePathFromNameAndType(params.name, EntityType.MetricsDefinition)
    );
  } catch (err) {
    if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
      throw error(404, "Dashboard not found");
    }

    throw error(err.response?.status || 500, err.message);
  }

  try {
    await runtimeServiceGetCatalogEntry(localConfig.instance_id, params.name);

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
