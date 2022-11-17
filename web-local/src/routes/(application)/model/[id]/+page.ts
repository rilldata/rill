import { browser } from "$app/environment";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { dataModelerStateService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { entityExists } from "@rilldata/web-local/lib/util/entity-exists";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  let modelExists = true;
  if (browser) {
    modelExists = await entityExists(
      dataModelerStateService.getEntityStateService(
        EntityType.Model,
        StateType.Persistent
      ).store,
      params.id
    );
  }

  if (modelExists) {
    return {
      modelID: params.id,
    };
  }

  throw error(404, "Model not found");
}
