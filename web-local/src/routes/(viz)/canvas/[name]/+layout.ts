import { handleCanvasStoreInitialization } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { LayoutLoad } from "./$types.js";

export const load: LayoutLoad = async ({ params }) => {
  return await handleCanvasStoreInitialization(params.name);
};
