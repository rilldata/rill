/**
 * A password-protected public URL grants no access until it is unlocked with its password,
 * which returns a session token to use as the bearer token in its place (see `UnlockMagicAuthToken`).
 *
 * The session is kept in sessionStorage, keyed by the public URL's own token, so it lasts for the
 * browser tab and is never written to the URL.
 */
const KEY_PREFIX = "rill:public-url-session:";

export interface PublicURLSession {
  token: string;
  expiresOn: string;
}

export function getPublicURLSession(
  token: string,
): PublicURLSession | undefined {
  if (typeof sessionStorage === "undefined") return undefined;
  const raw = sessionStorage.getItem(KEY_PREFIX + token);
  if (!raw) return undefined;
  try {
    const session = JSON.parse(raw) as PublicURLSession;
    if (session.token && new Date(session.expiresOn).getTime() > Date.now()) {
      return session;
    }
  } catch {
    // Unreadable entry: discard it below.
  }
  clearPublicURLSession(token);
  return undefined;
}

export function setPublicURLSession(token: string, session: PublicURLSession) {
  sessionStorage.setItem(KEY_PREFIX + token, JSON.stringify(session));
}

export function clearPublicURLSession(token: string) {
  sessionStorage.removeItem(KEY_PREFIX + token);
}
