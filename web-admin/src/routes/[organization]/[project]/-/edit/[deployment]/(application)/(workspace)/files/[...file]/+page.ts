import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { error, redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params: { file }, parent }) => {
  const parentData = await parent();

  if (!parentData.initialized) {
    console.log("Project not initialized, redirecting to home");
    throw redirect(303, "/");
  }

  // const readOnly = get(featureFlags.readOnly);
  const path = addLeadingSlash(file);

  // if (readOnly) {
  //   console.log("Read only, redirecting to home");
  //   throw redirect(303, "/");
  // }

  const fileArtifact = fileArtifacts.getFileArtifact(path);

  if (fileArtifact.fileTypeUnsupported) {
    throw error(
      400,
      "No renderer available for file type: " + fileArtifact.fileExtension,
    );
  }

  try {
    await fileArtifact.fetchContent();

    return {
      fileArtifact,
    };
  } catch (e) {
    const statusCode = e.response.status;

    if (statusCode === 404 || statusCode === 400) {
      throw error(404, "File not found: " + path);
    } else {
      throw error(e.response.status, e.response.data?.message);
    }
  }
};
