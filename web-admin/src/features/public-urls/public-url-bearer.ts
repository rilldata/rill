import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { extractErrorStatusCode } from "@rilldata/web-common/lib/errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getAdminServiceGetMagicAuthTokenQueryKey } from "./get-magic-auth-token";
import { clearPublicURLSession, getPublicURLSession } from "./password-session";

export interface PublicURLBearer {
  /**
   * The token to authenticate the public URL's requests with:
   * the session token once the URL has been unlocked, otherwise the URL's own token.
   */
  bearer: string;
  /** Whether the URL is password-protected and hasn't been unlocked in this tab yet. */
  locked: boolean;
  tokenMetadata: V1MagicAuthToken | undefined;
}

// The root loader resolves the bearer on every run, including each change to a dashboard's URL state.
// A token's metadata rarely changes, and the project layout notices on its own when the bearer stops working
// (see `invalidatePublicURLMetadata`), so a short cache spares a request on most runs.
const TOKEN_METADATA_STALE_TIME = 60 * 1000;

/**
 * Resolves the bearer token for a public URL.
 *
 * A password-protected token reveals nothing and grants no access until it is unlocked with its
 * password, which returns a session token to use in its place (see `PublicURLPasswordGate`).
 * A stored session that has expired or been invalidated (e.g. the password changed) is rejected
 * with a 401 and discarded, which locks the URL again. Any other failure (a network error,
 * a rate limit or a server error) is thrown and leaves the session in place.
 */
export async function resolvePublicURLBearer(
  token: string,
): Promise<PublicURLBearer> {
  const session = getPublicURLSession(token);
  if (session) {
    try {
      const { token: tokenMetadata } = await fetchMagicAuthToken(
        session.token,
        TOKEN_METADATA_STALE_TIME,
      );
      return { bearer: session.token, locked: false, tokenMetadata };
    } catch (e) {
      if (extractErrorStatusCode(e) !== 401) throw e;
    }
    clearPublicURLSession(token);
  }

  const { token: tokenMetadata, locked } = await fetchMagicAuthToken(
    token,
    TOKEN_METADATA_STALE_TIME,
  );
  return { bearer: token, locked: !!locked, tokenMetadata };
}

/**
 * Marks the cached metadata of a public URL's token as stale, so the next `resolvePublicURLBearer`
 * fetches it again. Call it when a response shows that a password was added to or removed from the URL.
 */
export function invalidatePublicURLMetadata(token: string) {
  return queryClient.invalidateQueries({
    queryKey: getAdminServiceGetMagicAuthTokenQueryKey(token),
  });
}
