import { redirect } from "@sveltejs/kit";

export const load = async ({ url }) => {
  if (url.searchParams.has("org")) {
    return redirect("/deploy/fresh-deploy?org=" + url.searchParams.get("org"));
  }
};
