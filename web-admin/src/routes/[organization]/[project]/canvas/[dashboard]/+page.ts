import { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";

export const load = async ({ url, params, parent }) => {
  const canvasName = params.dashboard;
  const {
    project: { id },
  } = await parent();

  await CanvasEntity.handleCanvasRedirect({
    canvasName,
    searchParams: url.searchParams,
    pathname: url.pathname,
    projectId: id,
  });

  return { canvasName };
};
