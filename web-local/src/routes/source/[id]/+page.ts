import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { dataModelerStateService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { entityExists } from "@rilldata/web-local/lib/util/entity-exists";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  // TODO: Check to see if the sourceId exists server-side
  const sourceExists = await entityExists(
    dataModelerStateService.getEntityStateService(
      EntityType.Table,
      StateType.Persistent
    ).store,
    params.id
  );

  if (sourceExists) {
    return {
      sourceId: params.id,
    };
  }

  console.log("params", params);
  throw error(404, "Source not found");
}
