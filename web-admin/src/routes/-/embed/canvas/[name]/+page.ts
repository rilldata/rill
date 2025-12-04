import type { PageLoad } from "./$types.js";

export const load: PageLoad = async ({ url, parent }) => {
  const parentData = await parent();

  const { store, canvasName } = await parent();

  await store.canvasEntity.onUrlChange({ url, loadFunction: true });

  return { canvasName, parentData };
};
