import { goto } from "$app/navigation";
import { page } from "$app/stores";
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
import { get } from "svelte/store";

export async function checkUserAccess() {
  // Check for a logged-in user
  const userQuery = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
    queryKey: getAdminServiceGetCurrentUserQueryKey(),
    queryFn: () => adminServiceGetCurrentUser(),
  });
  const isLoggedIn = !!userQuery.user;

  const pageState = get(page);

  // If not logged in, redirect to the login page
  if (!isLoggedIn) {
    redirectToLogin();
    return true;
  } else if (
    withinProject(pageState) &&
    !isProjectRequestAccessPage(pageState)
  ) {
    // if not in request access page (approve or deny routes) then go to a page to get access
    await goto(
      `/-/request-project-access/?organization=${pageState.params.organization}&project=${pageState.params.project}`,
    );
    return true;
  }

  return false;
}
