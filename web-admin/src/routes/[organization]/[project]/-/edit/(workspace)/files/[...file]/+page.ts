import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
import { consumeViewSearchParam } from "@rilldata/web-common/layout/workspace/workspace-stores";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ params: { file }, parent, url }) => {
  await parent();
  const path = addLeadingSlash(file);

  const urlWithoutViewParam = consumeViewSearchParam(url, path);
  if (urlWithoutViewParam) {
    throw redirect(307, urlWithoutViewParam);
  }

  const fileArtifact = fileArtifacts.getFileArtifact(path);

  if (fileArtifact.fileTypeUnsupported) {
    throw error(
      400,
      "No renderer available for file type: " + fileArtifact.fileExtension,
    );
  }

  // Don't eagerly fetch content here. Unlike web-local, the runtime
  // credentials aren't available until the edit layout fetches them
  // asynchronously. The workspace components will fetch content
  // reactively once the runtime is ready.

  return {
    fileArtifact,
  };
};
