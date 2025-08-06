import { redirect } from "@sveltejs/kit";

export const load = ({ url: { searchParams } }) => {
  const orgName = searchParams.get("org");
  const projectName = searchParams.get("project");
  if (!orgName || !projectName) throw redirect(307, "/deploy");

  const createManagedRepo = searchParams.get("create_managed_repo") === "true";
  return {
    orgName,
    projectName,
    createManagedRepo,
  };
};
