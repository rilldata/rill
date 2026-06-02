import { VirtualFileIo } from "@rilldata/web-admin/features/virtual-file-editor/virtual-file-io.ts";

export const load = async ({ params: { organization, project }, parent }) => {
  const fileIo = new VirtualFileIo(organization, project);

  return { fileIo };
};
