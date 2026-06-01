import { generateBlobForNewResourceFile } from "@rilldata/web-common/features/entity-management/add/new-files.ts";
import { VirtualFileIo } from "@rilldata/web-admin/features/virtual-file-editor/virtual-file-io.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export const load = async ({ params: { organization, project }, parent }) => {
  const { user } = await parent();
  const fileIo = new VirtualFileIo(organization, project, user.id);

  return { fileIo };
};
