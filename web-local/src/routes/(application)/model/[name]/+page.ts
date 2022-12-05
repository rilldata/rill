import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const localConfig = await runtimeServiceGetConfig();

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
