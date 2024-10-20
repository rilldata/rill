import { redirect } from "@sveltejs/kit";

export const load = ({ params }) => {
  throw redirect(307, `/${params.organization}/${params.project}`);
};
