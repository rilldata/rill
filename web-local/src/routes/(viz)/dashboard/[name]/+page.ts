import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

/**
 * Redirect `/dashboard/[name]` to `/explore/[name]`.
 * Maintains backwards compatibility with legacy URLs.
 */
export const load: PageLoad = ({ params }) => {
  throw redirect(307, `/explore/${params.name}`);
};
