import { browser } from "$app/environment";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { dataModelerStateService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { entityExists } from "@rilldata/web-local/lib/util/entity-exists";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export const ssr = false;
export async function load({ params }) {
  let sourceExists = true;
  if (browser) {
    sourceExists = await entityExists(
      dataModelerStateService.getEntityStateService(
        EntityType.Table,
        StateType.Persistent
      ).store,
      params.id
    );
  }
  if (sourceExists) {
    return {
      sourceID: params.id,
    };
  }

  throw error(404, "Source not found");
}
