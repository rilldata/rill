import { error, redirect } from "@sveltejs/kit";
import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { get } from "svelte/store";

export const load = async ({ params: { file } }) => {
  const readOnly = get(featureFlags.readOnly);
  const path = addLeadingSlash(file);

  if (readOnly) {
    throw redirect(303, "/");
  }

  const fileArtifact = fileArtifacts.getFileArtifact(path);

  if (fileArtifact.fileTypeUnsupported) {
    throw error(400, fileArtifact.fileExtension + " file type not supported");
  }

  try {
    await fileArtifact.fetchContent();

    return {
      filePath: path,
      fileArtifact,
    };
  } catch (e) {
    throw error(404, "File not found: " + path);
  }
};
