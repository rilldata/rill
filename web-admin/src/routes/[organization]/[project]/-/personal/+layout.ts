import { VirtualFileIo } from "@rilldata/web-admin/features/personal-files/virtual-file-io.ts";

export const load = async ({ params: { organization, project }, parent }) => {
  const fileIo = new VirtualFileIo(organization, project);

  return { fileIo };
};
