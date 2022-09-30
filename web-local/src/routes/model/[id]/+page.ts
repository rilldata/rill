import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { dataModelerStateService } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { entityExists } from "@rilldata/web-local/lib/util/entity-exists";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const modelExists = await entityExists(
    dataModelerStateService.getEntityStateService(
      EntityType.Model,
      StateType.Persistent
    ).store,
    params.id
  );

  if (modelExists) {
    return {
      modelId: params.id,
    };
  }

  throw error(404, "Model not found");
}
