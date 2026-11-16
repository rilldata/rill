/**
 * `UnlockMagicAuthToken` must be called with the public URL's token as the bearer token,
 * so it can't use the generated client, which relies on cookie-based auth (see `get-magic-auth-token.ts`).
 */
import type { V1UnlockMagicAuthTokenResponse } from "@rilldata/web-admin/client";
import httpClient from "@rilldata/web-admin/client/http-client";

export const adminServiceUnlockMagicAuthToken = (
  token: string,
  password: string,
) => {
  return httpClient<V1UnlockMagicAuthTokenResponse>({
    url: `/v1/magic-tokens/current/unlock`,
    method: "post",
    data: { password },
    headers: {
      Authorization: `Bearer ${token}`,
    },
    // There's no need to send credentials (cookies) with this request
    withCredentials: false,
  });
};
