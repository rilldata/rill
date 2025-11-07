import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import { redirect } from "@sveltejs/kit";

export const load = async ({ url, params }) => {
  const urlSearchParams = url.searchParams.toString();

  if (url.searchParams.get("home") !== null) {
    throw redirect(307, url.pathname);
  }
  const canvasName = params.dashboard;
  const snapshotSearchParams = lastVisitedState.get(canvasName);

  if (snapshotSearchParams && !urlSearchParams) {
    throw redirect(307, `?${snapshotSearchParams}`);
  }
};
