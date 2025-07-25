import { redirect } from "@sveltejs/kit";

export const load = ({ url: { searchParams } }) => {
  const orgName = searchParams.get("org");
  const projectName = searchParams.get("project");
  if (!orgName || !projectName) throw redirect(307, "/deploy");

  const newManagedRepo = searchParams.get("new_managed_repo") ?? false;
  return {
    orgName,
    projectName,
    newManagedRepo,
  };
};
