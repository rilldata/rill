import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { redirect } from "@sveltejs/kit";

export const load = async ({ url, params }) => {
  const canvasName = params.dashboard;
  const snapshotSearchParams = lastVisitedState.get(canvasName);

  if (snapshotSearchParams && !url.search.toString()) {
    throw redirect(307, `?${snapshotSearchParams}`);
  }
};
