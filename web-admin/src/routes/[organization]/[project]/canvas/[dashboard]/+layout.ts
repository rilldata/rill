import { handleCanvasStoreInitialization } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { LayoutLoad } from "./$types.js";

export const load: LayoutLoad = async ({ params, parent }) => {
  const canvasName = params.dashboard;

  const {
    runtime: { instanceId },
  } = await parent();

  return await handleCanvasStoreInitialization(canvasName, instanceId);
};
