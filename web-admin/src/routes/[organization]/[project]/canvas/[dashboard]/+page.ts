import type { PageLoad } from "./$types.js";

export const load: PageLoad = async ({ url, parent }) => {
  const {
    store,
    canvasName,
    project: { id: projectId },
  } = await parent();

  await store.canvasEntity.onUrlChange({ url, loadFunction: true, projectId });

  return { canvasName, store };
};
