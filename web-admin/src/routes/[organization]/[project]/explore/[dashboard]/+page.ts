import {
  fetchBookmarks,
  isHomeBookmark,
} from "@rilldata/web-admin/features/bookmarks/selectors";

export const load = async ({ params, parent }) => {
  const { user, project } = await parent();
  const { dashboard: exploreName } = params;

  if (user) {
    try {
      const bookmarks = await fetchBookmarks(project.id, exploreName);
      return {
        // We are returning just the bookmark and not the parsed explore state because we need schema that queries the datastore.
        // This can take a while in certain cases, since there is no way to handle the loading state from here we push it to components.
        homeBookmark: bookmarks.find(isHomeBookmark),
      };
    } catch {
      // no-op
    }
  }

  return {
    homeBookmark: undefined,
  };
};
