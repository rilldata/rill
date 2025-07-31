import { redirect } from "@sveltejs/kit";

export const load = ({ url: { searchParams } }) => {
  const org = searchParams.get("org");
  const projectId = searchParams.get("project_id");
  if (!org || !projectId) throw redirect(307, "/deploy");
  return { org, projectId };
};
