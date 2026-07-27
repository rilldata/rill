import { VirtualFileIo } from "@rilldata/web-admin/features/personal-files/virtual-file-io.ts";

export const load = async ({ params: { organization, project } }) => {
  const fileIo = new VirtualFileIo(organization, project);

  return { fileIo };
};
