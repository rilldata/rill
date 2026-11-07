import type { V1MagicAuthToken } from "@rilldata/web-admin/client";
import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
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

/**
 * Resolves the bearer token for a public URL.
 *
 * A password-protected token reveals nothing and grants no access until it is unlocked with its
 * password, which returns a session token to use in its place (see `PublicURLPasswordGate`).
 * A stored session that has expired or been invalidated (e.g. the password changed) is discarded,
 * which locks the URL again.
 */
export async function resolvePublicURLBearer(
  token: string,
): Promise<PublicURLBearer> {
  const session = getPublicURLSession(token);
  if (session) {
    try {
      const { token: tokenMetadata, locked } = await fetchMagicAuthToken(
        session.token,
      );
      if (!locked) {
        return { bearer: session.token, locked: false, tokenMetadata };
      }
    } catch (e) {
      console.warn(e);
    }
    clearPublicURLSession(token);
  }

  const { token: tokenMetadata, locked } = await fetchMagicAuthToken(token);
  return { bearer: token, locked: !!locked, tokenMetadata };
}
