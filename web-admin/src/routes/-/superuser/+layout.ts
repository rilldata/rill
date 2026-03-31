// Superuser console layout guard: verifies the current user is a superuser
import {
  adminServiceGetCurrentUser,
  getAdminServiceGetCurrentUserQueryKey,
  type V1GetCurrentUserResponse,
} from "@rilldata/web-admin/client";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async () => {
  let currentUserEmail: string | undefined;
  let isSuperuser = false;
  try {
    const userResp = await queryClient.fetchQuery<V1GetCurrentUserResponse>({
      queryKey: getAdminServiceGetCurrentUserQueryKey(),
      queryFn: () => adminServiceGetCurrentUser(),
      staleTime: 5 * 60 * 1000,
    });
    currentUserEmail = userResp.user?.email;
    isSuperuser = userResp.superuser ?? false;
  } catch (e) {
    if (isAxiosError(e) && e.response?.status === 401) {
      redirectToLogin();
    }
    throw redirect(307, "/");
  }

  if (!currentUserEmail || !isSuperuser) {
    throw redirect(307, "/");
  }

  return { currentUserEmail };
};
