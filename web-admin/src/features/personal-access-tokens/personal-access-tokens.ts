import {
  adminServiceIssueUserAuthToken,
  adminServiceListUserAuthTokens,
  getAdminServiceListUserAuthTokensQueryKey,
  type V1UserAuthToken,
} from "@rilldata/web-admin/client";
import { getLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
import { createQuery, type QueryClient } from "@tanstack/svelte-query";
import { DateTime } from "luxon";

// Personal access tokens are the user tokens issued under the "Created manually" auth client.
// Keep in sync with AuthClientIDRillManual in admin/database/database.go.
export const PERSONAL_ACCESS_TOKEN_CLIENT_ID =
  "12345678-0000-0000-0000-000000000005";

export const ACCOUNT_TOKENS_PATH = "/-/account/tokens";

export const TOKEN_NAME_MAX_LENGTH = 64;

// Tokens that expire within this many days are highlighted in the list.
export const EXPIRES_SOON_DAYS = 7;

export type TokenExpiration = "7d" | "30d" | "90d" | "1y" | "custom" | "none";

export const DEFAULT_TOKEN_EXPIRATION: TokenExpiration = "90d";

const EXPIRATION_PRESET_DAYS: Record<
  Exclude<TokenExpiration, "custom" | "none">,
  number
> = {
  "7d": 7,
  "30d": 30,
  "90d": 90,
  "1y": 365,
};

/**
 * Returns when a token issued at `now` with the given expiration would expire, or null if it never expires.
 * A custom date keeps the current time of day, so a token picked for tomorrow lasts a full day.
 */
export function getTokenExpiryDate(
  expiration: TokenExpiration,
  customDate: DateTime | null,
  now: DateTime = DateTime.now(),
): DateTime | null {
  switch (expiration) {
    case "none":
      return null;
    case "custom":
      if (!customDate) return null;
      return customDate.set({
        hour: now.hour,
        minute: now.minute,
        second: now.second,
        millisecond: now.millisecond,
      });
    default:
      return now.plus({ days: EXPIRATION_PRESET_DAYS[expiration] });
  }
}

/**
 * Converts an expiry date to the `ttl_minutes` of IssueUserAuthToken, where 0 means the token never expires.
 */
export function getTokenTTLMinutes(
  expiresOn: DateTime | null,
  now: DateTime = DateTime.now(),
): number {
  if (!expiresOn) return 0;
  return Math.max(1, Math.round(expiresOn.diff(now, "minutes").minutes));
}

/**
 * Issues a personal access token for the current user and refreshes the token list.
 * It calls the API function rather than a mutation, so the secret stays out of TanStack's mutation cache.
 */
export async function issuePersonalAccessToken(
  queryClient: QueryClient,
  displayName: string,
  expiration: TokenExpiration,
  customDate: DateTime | null,
): Promise<{ token: string; expiresOn: DateTime | null }> {
  const now = DateTime.now();
  const expiresOn = getTokenExpiryDate(expiration, customDate, now);
  const resp = await adminServiceIssueUserAuthToken("current", {
    clientId: PERSONAL_ACCESS_TOKEN_CLIENT_ID,
    displayName,
    ttlMinutes: String(getTokenTTLMinutes(expiresOn, now)),
  });
  void queryClient.invalidateQueries({
    queryKey: getAdminServiceListUserAuthTokensQueryKey("current"),
  });
  return { token: resp.token ?? "", expiresOn };
}

/**
 * Keeps the user's own personal access tokens, newest first.
 * The list endpoint also returns browser sessions, CLI and Rill Developer logins,
 * OAuth refresh tokens and support tokens, which are not personal access tokens.
 */
export function selectPersonalAccessTokens(
  tokens: V1UserAuthToken[],
): V1UserAuthToken[] {
  return tokens
    .filter(
      (t) =>
        t.authClientId === PERSONAL_ACCESS_TOKEN_CLIENT_ID &&
        !t.representingUserId &&
        !t.refresh,
    )
    .sort(
      (a, b) =>
        Date.parse(b.createdOn ?? "") - Date.parse(a.createdOn ?? "") || 0,
    );
}

/**
 * Returns when the token was last used, or undefined if it never was.
 * The admin database sets `used_on` to the creation time when it issues a token.
 */
export function getTokenLastUsedOn(token: V1UserAuthToken): string | undefined {
  if (!token.usedOn) return undefined;
  if (Date.parse(token.usedOn) === Date.parse(token.createdOn ?? "")) {
    return undefined;
  }
  return token.usedOn;
}

export function isExpiringSoon(
  expiresOn: string | undefined,
  now: DateTime = DateTime.now(),
): boolean {
  if (!expiresOn) return false;
  return DateTime.fromISO(expiresOn) <= now.plus({ days: EXPIRES_SOON_DAYS });
}

export function formatTokenDate(date: DateTime | string): string {
  const value = typeof date === "string" ? new Date(date) : date.toJSDate();
  return value.toLocaleDateString(getLocale(), {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

/**
 * Lists the current user's personal access tokens across all pages.
 * Invalidate it with `getAdminServiceListUserAuthTokensQueryKey("current")`.
 */
export function createPersonalAccessTokensQuery() {
  return createQuery({
    queryKey: [
      ...getAdminServiceListUserAuthTokensQueryKey("current"),
      "personal-access-tokens",
    ],
    queryFn: async ({ signal }) => {
      const tokens: V1UserAuthToken[] = [];
      let pageToken: string | undefined;
      do {
        const resp = await adminServiceListUserAuthTokens(
          "current",
          { pageSize: 1000, pageToken, refresh: false },
          signal,
        );
        tokens.push(...(resp.tokens ?? []));
        pageToken = resp.nextPageToken || undefined;
      } while (pageToken);
      return selectPersonalAccessTokens(tokens);
    },
  });
}
