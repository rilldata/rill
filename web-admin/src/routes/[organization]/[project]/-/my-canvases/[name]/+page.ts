import { adminServiceGetPersonalVirtualFile } from "@rilldata/web-admin/client";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params, url }) => {
  const { organization, project, name } = params;
  const mode = url.searchParams.get("mode") === "edit" ? "edit" : "view";

  try {
    const data = await adminServiceGetPersonalVirtualFile(
      organization,
      project,
      "PERSONAL_VIRTUAL_FILE_TYPE_CANVAS",
      name,
    );
    return {
      canvasName: name,
      displayName: data.displayName ?? name,
      yaml: data.yaml ?? "",
      mode,
    };
  } catch (e) {
    // Either the canvas does not exist or the caller is not the owner. Either way: 404.
    throw error(404, "Personal canvas not found");
  }
};
