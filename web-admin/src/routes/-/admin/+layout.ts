// web-admin/src/routes/-/admin/+layout.ts
import {
  adminServiceGetCurrentUser,
  adminServiceListSuperusers,
  getAdminServiceGetCurrentUserQueryKey,
  getAdminServiceListSuperusersQueryKey,
  type V1GetCurrentUserResponse,
  type V1ListSuperusersResponse,
} from "@rilldata/web-admin/client";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async () => {
  // Get current user
  let currentUserEmail: string | undefined;
  try {
    const userResp = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
      queryKey: getAdminServiceGetCurrentUserQueryKey(),
      queryFn: () => adminServiceGetCurrentUser(),
      staleTime: 5 * 60 * 1000,
    });
    currentUserEmail = userResp.user?.email;
  } catch (e) {
    if (isAxiosError(e) && e.response?.status === 401) {
      // redirectToLogin() throws a SvelteKit redirect internally;
      // call it outside the catch to avoid swallowing the redirect exception
    } else {
      throw redirect(307, "/");
    }
    redirectToLogin();
  }

  if (!currentUserEmail) {
    throw redirect(307, "/");
  }

  // Check if current user is a superuser
  try {
    const superusersResp =
      await queryClient.fetchQuery<V1ListSuperusersResponse>({
        queryKey: getAdminServiceListSuperusersQueryKey(),
        queryFn: () => adminServiceListSuperusers(),
        staleTime: 5 * 60 * 1000,
      });

    const isSuperuser = superusersResp.users?.some(
      (u) => u.email === currentUserEmail,
    );

    if (!isSuperuser) {
      throw redirect(307, "/");
    }
  } catch (e) {
    // ListSuperusers itself will 403 if not a superuser
    if (isAxiosError(e) && e.response?.status === 403) {
      throw redirect(307, "/");
    }
    // Re-throw SvelteKit redirects
    throw e;
  }

  return { currentUserEmail };
};
