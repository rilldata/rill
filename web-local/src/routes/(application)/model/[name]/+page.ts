import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { EntityType } from "@rilldata/web-common/features/entity-management/types";
import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params, url }) {
  /** If ?focus, tell the page to focus the editor as soon as available */
  const focusEditor = url.searchParams.get("focus") === "";
  try {
    const config = await runtimeServiceGetConfig();
    if (config.readonly) {
      throw error(404, "Page not found");
    }

    await runtimeServiceGetFile(
      config.instance_id,
      getFilePathFromNameAndType(params.name, EntityType.Model)
    );

    return {
      modelName: params.name,
      focusEditor,
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
