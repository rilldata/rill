import { redirect } from "@sveltejs/kit";

export const load = ({ url: { searchParams } }) => {
  const org = searchParams.get("org");
  if (!org) throw redirect(307, "/deploy");
  return { org, useGit: searchParams.get("use_git") === "true" };
};
