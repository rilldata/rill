import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params }) => {
  const { organization } = params;

  return {
    organization,
  };
};