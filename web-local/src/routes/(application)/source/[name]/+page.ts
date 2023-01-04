import { EntityType } from "@rilldata/web-common/lib/entity";
import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const localConfig = await runtimeServiceGetConfig();

    await runtimeServiceGetFile(
      localConfig.instance_id,
      getFilePathFromNameAndType(params.name, EntityType.Table)
    );

    return {
      sourceName: params.name,
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
