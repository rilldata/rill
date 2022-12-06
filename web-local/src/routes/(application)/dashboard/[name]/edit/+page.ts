import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import {
  ExplorerSourceColumnDoesntExist,
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
  ExplorerTimeDimensionDoesntExist,
} from "@rilldata/web-local/common/errors/ErrorMessages";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await runtimeServiceGetConfig();

  try {
    await runtimeServiceGetFile(
      localConfig.instance_id,
      `dashboards/${params.name}.yaml`
    );

    return {
      metricsDefName: params.name,
    };
  } catch (err) {
    if (err.response?.data?.message.includes("entry not found")) {
      throw error(404, "Dashboard not found");
    }

    // The following invalid dashboard errors are displayed by the component
    const invalidDashboardErrors = [
      ExplorerSourceModelDoesntExist,
      ExplorerSourceModelIsInvalid,
      ExplorerSourceColumnDoesntExist,
      ExplorerTimeDimensionDoesntExist,
    ];
    if (
      invalidDashboardErrors.some(
        (errMsg) => errMsg.includes(err.message) || err.message.includes(errMsg)
      )
    ) {
      return {
        metricsName: params.name,
        error: err.message,
      };
    }

    // Throw all other errors
    throw error(err.response?.status || 500, err.message);
  }
}
