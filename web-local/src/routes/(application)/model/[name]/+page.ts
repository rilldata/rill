import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { error } from "@sveltejs/kit";
import { fetchWrapperDirect } from "../../../../lib/util/fetchWrapper";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const localConfig = await fetchWrapperDirect(
      `${config.server.serverUrl}/local/config`,
      "GET"
    );

    await runtimeServiceGetFile(
      localConfig.instance_id,
      `models/${params.name}.sql`
    );

    return {
      modelName: params.name,
    };
  } catch (e) {
    if (e.response?.status && e.response?.data?.message) {
      throw error(e.response.status, e.response.data.message);
    } else {
      console.error(e);
      throw error(500, e.message);
    }
  }
}
