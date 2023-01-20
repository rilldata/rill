import { EntityType } from "@rilldata/web-common/lib/entity";
import {
  runtimeServiceGetCatalogEntry,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import { error } from "@sveltejs/kit";
import { parseDocument } from "yaml";
import { CATALOG_ENTRY_NOT_FOUND } from "../../../../lib/errors/messages";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await runtimeServiceGetConfig();
  let file;
  try {
    file = await runtimeServiceGetFile(
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
    const { entry } = await runtimeServiceGetCatalogEntry(
      localConfig.instance_id,
      params.name
    );

    return {
      configName: params.name,
      entry,
      validDashboard: true,
      modelExists: true,
    };
  } catch (err) {
    // file not in catalog, return file itself.
    if (err?.response?.status === 400) {
      try {
        const metricsView = parseDocument(file.blob || "{}").toJS();

        // const modelDefined = metricsView?.model?.length > 0;
        let modelExists = false;
        try {
          await runtimeServiceGetCatalogEntry(
            localConfig.instance_id,
            metricsView?.model
          );
          modelExists = true;
        } catch (err) {
          // no-op
        }

        return {
          configName: params.name,
          entry: { metricsView },
          validDashboard: false,
          modelExists,
        };
      } catch (err) {
        throw error(400, "Invalid metrics definition");
      }
    } else {
      throw error(err.response?.status || 500, err.message);
    }
    console.log(file);
    // If the catalog entry doesn't exist, the dashboard config is invalid, so we redirect to the dashboard editor
    console.log("lets figure this one out", err);
    //throw redirect(307, `/dashboard/${params.name}/edit`);
  }
}
