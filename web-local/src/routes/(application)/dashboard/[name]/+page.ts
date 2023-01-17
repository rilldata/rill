import { EntityType } from "@rilldata/web-common/lib/entity";
import {
  runtimeServiceGetCatalogEntry,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import { error, redirect } from "@sveltejs/kit";
import { CATALOG_ENTRY_NOT_FOUND } from "../../../../lib/errors/messages";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const config = await runtimeServiceGetConfig();

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
      metricViewName: params.name,
    };
  } catch (err) {
    // If the catalog entry doesn't exist, the dashboard config is invalid, so we redirect to the dashboard editor
    throw redirect(307, `/dashboard/${params.name}/edit`);
  }
}
