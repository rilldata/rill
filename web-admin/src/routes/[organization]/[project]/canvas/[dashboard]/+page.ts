// import type { PageLoad } from './$types';
// import { getCanvasCategorisedBookmarks } from '@rilldata/web-admin/features/bookmarks/selectors';

export const load = async ({ params }) => {
  const { organization, project, dashboard } = params;

  // const what = await getCanvasCategorisedBookmarks(organization, project, dashboard);
  return {
    organization,
    project,
    dashboard,
  };
};
