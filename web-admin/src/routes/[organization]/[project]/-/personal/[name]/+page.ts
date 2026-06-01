import { generateBlobForNewResourceFile } from "@rilldata/web-common/features/entity-management/add/new-files.ts";
import { VirtualFileIo } from "@rilldata/web-admin/features/virtual-file-editor/virtual-file-io.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export const load = async ({ params: { name }, parent }) => {
  const { fileIo } = await parent();

  const path = `/personal/canvas_${user.id}`;
  const contents = await fileIo.read(path, false);
  if (!contents) {
    await fileIo.write(
      path,
      generateBlobForNewResourceFile(ResourceKind.Canvas),
      ResourceKind.Canvas,
    );
  }
};
