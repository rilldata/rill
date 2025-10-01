import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params }) => {
  const { organization, project } = params;
  redirect(307, `/${organization}/${project}/-/chat`);
};
