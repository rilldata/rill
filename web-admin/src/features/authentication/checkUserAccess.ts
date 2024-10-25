import {
  adminServiceGetCurrentUser,
  getAdminServiceGetCurrentUserQueryKey,
  type V1GetCurrentUserResponse,
} from "@rilldata/web-admin/client";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

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
