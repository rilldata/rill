import { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";

export const load = async ({ url, params }) => {
  const canvasName = params.name;

  await CanvasEntity.handleCanvasRedirect({
    canvasName,
    searchParams: url.searchParams,
    pathname: url.pathname,
  });

  return { canvasName };
};
