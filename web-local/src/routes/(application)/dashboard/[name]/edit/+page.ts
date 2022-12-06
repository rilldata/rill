import {
  runtimeServiceGetCatalogEntry,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await runtimeServiceGetConfig();

  try {
    await runtimeServiceGetFile(
      localConfig.instance_id,
      `dashboards/${params.name}.yaml`
    );
  } catch (err) {
    if (err.response?.data?.message.includes("entry not found")) {
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
    if (err.response?.data?.message.includes("entry not found")) {
      return {
        metricsDefName: params.name,
        error: err.message,
      };
    }

    // Throw all other errors
    throw error(err.response?.status || 500, err.message);
  }
}
