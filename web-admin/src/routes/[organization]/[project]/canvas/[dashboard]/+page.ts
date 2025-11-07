import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { redirect } from "@sveltejs/kit";

export const load = async ({ url: { searchParams }, params }) => {
  const urlSearchParams = searchParams.toString();
  const canvasName = params.dashboard;
  const snapshotSearchParams = lastVisitedState.get(canvasName);

  if (snapshotSearchParams && !urlSearchParams) {
    throw redirect(307, `?${snapshotSearchParams}`);
  }
};
