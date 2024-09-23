import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

/**
 * Previously, Explores were located at `/[organization]/[project]/[dashboard]`.
 * Now we redirect this route to `/[organization]/[project]/explore/[dashboard]`.
 */
export const load: PageLoad = ({ params }) => {
  throw redirect(
    307,
    `/${params.organization}/${params.project}/explore/${params.dashboard}`,
  );
};
