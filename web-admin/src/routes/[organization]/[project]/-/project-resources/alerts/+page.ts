import { redirect } from "@sveltejs/kit";

export const load = ({ params }) => {
  throw redirect(301, `/${params.organization}/${params.project}/-/alerts`);
};
