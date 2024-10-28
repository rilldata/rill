import { goto } from "$app/navigation";
import {
  adminServiceGetCurrentUser,
  getAdminServiceGetCurrentUserQueryKey,
  type V1GetCurrentUserResponse,
} from "@rilldata/web-admin/client";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import {
  isProjectRequestAccessPage,
  withinProject,
} from "@rilldata/web-admin/features/navigation/nav-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { Page } from "@sveltejs/kit";

export async function redirectToLoginIfNotLoggedIn() {
  // Check for a logged-in user
  const userQuery = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
    queryKey: getAdminServiceGetCurrentUserQueryKey(),
    queryFn: () => adminServiceGetCurrentUser(),
  });
  const isLoggedIn = !!userQuery.user;
  if (isLoggedIn) {
    return false;
  }

  // If not logged in, redirect to the login page
  redirectToLogin();
  return true;
}

export async function redirectToLoginOrRequestAccess(page: Page) {
  const didRedirect = await redirectToLoginIfNotLoggedIn();
  if (didRedirect) return true;
  if (withinProject(page) && !isProjectRequestAccessPage(page)) {
    // if not in request access page (approve or deny routes) then go to a page to get access
    await goto(
      `/-/request-project-access/?organization=${page.params.organization}&project=${page.params.project}`,
    );
    return true;
  }
  return false;
}
