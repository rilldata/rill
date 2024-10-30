import { redirect } from "@sveltejs/kit";

/**
 * Redirect `/[organization]/[project]/[dashboard]` to `/[organization]/[project]/explore/[dashboard]`.
 * Maintains backwards compatibility with legacy URLs.
 */
export const load = ({ params }) => {
  throw redirect(
    307,
    `/${params.organization}/${params.project}/explore/${params.dashboard}`,
  );
};
