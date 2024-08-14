import {
  adminServiceGetMagicAuthToken,
  getAdminServiceGetMagicAuthTokenQueryKey,
} from "@rilldata/web-admin/features/public-urls/get-magic-auth-token";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { error } from "@sveltejs/kit";
import { type QueryFunction } from "@tanstack/svelte-query";

export const load = async ({ params: { token }, url: { searchParams } }) => {
  const queryKey = getAdminServiceGetMagicAuthTokenQueryKey(token);
  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>
  > = ({ signal }) => adminServiceGetMagicAuthToken(token, signal);

  try {
    const tokenData = await queryClient.fetchQuery({
      queryKey,
      queryFn: queryFunction,
    });

    const state = searchParams.get("state");

    // Add the token's `state` to the URL (only if there's no existing URL `state`)
    if (tokenData?.token?.state && !state) {
      searchParams.set("state", tokenData.token.state);
    }

    return {
      token: tokenData?.token,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find token");
  }
};
