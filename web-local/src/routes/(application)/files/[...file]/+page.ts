import { error } from "@sveltejs/kit";
import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";

export const load = async ({ params: { file } }) => {
  const path = addLeadingSlash(file);

  const fileArtifact = fileArtifacts.getFileArtifact(path);

  if (fileArtifact.fileTypeUnsupported) {
    throw error(400, fileArtifact.fileExtension + " file type not supported");
  }

  try {
    await fileArtifact.initRemoteContent();

    return {
      filePath: path,
      fileArtifact,
    };
  } catch (e) {
    throw error(404, "File not found: " + path);
  }
};
