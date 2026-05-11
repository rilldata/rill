import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { file } }) => {
  const path = addLeadingSlash(file);
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
