import { EntityType } from "@rilldata/web-common/lib/entity";
import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
import { error, redirect } from "@sveltejs/kit";
import { parseDocument } from "yaml";
import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await runtimeServiceGetConfig();

  // first, we will get the model name out of the YAML file.
  // It's not critical that the config is a fully-valid config; just
  // that it parses & model has been defined in it.
  let configFile;
  try {
    configFile = await runtimeServiceGetFile(
      localConfig.instance_id,
      getFilePathFromNameAndType(params.name, EntityType.MetricsDefinition)
    );
  } catch (err) {
    if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
      throw error(404, "Dashboard not found");
    }
    throw error(err.response?.status || 500, err.message);
  }
  const parsedConfig = parseDocument(configFile.blob || "{}").toJS();

  try {
    // check to see if the file exists.
    // if it does, return the model name.
    await runtimeServiceGetFile(
      localConfig.instance_id,
      getFilePathFromNameAndType(parsedConfig.model, EntityType.Model)
    );

    return {
      modelName: parsedConfig.model,
    };
  } catch (err) {
    // If the catalog entry doesn't exist, the config is invalid,
    // so we redirect to the configuration editor.
    throw redirect(307, `/dashboard/${params.name}/edit`);
  }
}
