import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ params: { file }, parent }) => {
  const parentData = await parent();

  if (!parentData.initialized) {
    throw redirect(303, "/");
  }

  const path = addLeadingSlash(file);

  // Only allow CSV files in the data folder
  if (!path.startsWith("/data/") || !path.endsWith(".csv")) {
    throw error(400, "Only CSV files in the data folder can be edited");
  }

  const fileArtifact = fileArtifacts.getFileArtifact(path);

  try {
    await fileArtifact.fetchContent();

    return {
      fileArtifact,
      filePath: path,
    };
  } catch (e) {
    const statusCode = e.response?.status;

    if (statusCode === 404 || statusCode === 400) {
      throw error(404, "File not found: " + path);
    } else {
      throw error(e.response?.status ?? 500, e.response?.data?.message ?? "Unknown error");
    }
  }
};
