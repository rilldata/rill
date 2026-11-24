import type { V1UserAuthToken } from "@rilldata/web-admin/client";
import { DateTime } from "luxon";
import { describe, expect, it } from "vitest";
import {
  getTokenExpiryDate,
  getTokenLastUsedOn,
  getTokenTTLMinutes,
  isExpiringSoon,
  PERSONAL_ACCESS_TOKEN_CLIENT_ID,
  selectPersonalAccessTokens,
  type TokenExpiration,
} from "./personal-access-tokens";

const now = DateTime.fromISO("2026-11-25T14:30:15.250Z", { zone: "utc" });

describe("token expiration", () => {
  it.each<[TokenExpiration, number]>([
    ["7d", 7 * 24 * 60],
    ["30d", 30 * 24 * 60],
    ["90d", 90 * 24 * 60],
    ["1y", 365 * 24 * 60],
  ])("issues %s as %i minutes", (expiration, minutes) => {
    const expiresOn = getTokenExpiryDate(expiration, null, now);
    expect(getTokenTTLMinutes(expiresOn, now)).toBe(minutes);
  });

  it("issues no expiration as 0 minutes", () => {
    const expiresOn = getTokenExpiryDate("none", null, now);
    expect(expiresOn).toBeNull();
    expect(getTokenTTLMinutes(expiresOn, now)).toBe(0);
  });

  it("keeps the current time of day for a custom date", () => {
    const tomorrow = now.plus({ days: 1 }).startOf("day");
    const expiresOn = getTokenExpiryDate("custom", tomorrow, now);
    expect(expiresOn?.toISO()).toBe("2026-11-26T14:30:15.250Z");
    expect(getTokenTTLMinutes(expiresOn, now)).toBe(24 * 60);
  });
});

describe("selectPersonalAccessTokens", () => {
  const token = (
    id: string,
    createdOn: string,
    extra: Partial<V1UserAuthToken> = {},
  ): V1UserAuthToken => ({
    id,
    createdOn,
    authClientId: PERSONAL_ACCESS_TOKEN_CLIENT_ID,
    ...extra,
  });

  it("keeps only the user's own manually created tokens, newest first", () => {
    const tokens = [
      token("old", "2026-01-01T00:00:00Z"),
      token("browser", "2026-11-24T00:00:00Z", {
        authClientId: "12345678-0000-0000-0000-000000000001",
      }),
      token("cli", "2026-11-24T00:00:00Z", {
        authClientId: "12345678-0000-0000-0000-000000000002",
      }),
      token("assumed", "2026-11-24T00:00:00Z", { representingUserId: "u2" }),
      token("refresh", "2026-11-24T00:00:00Z", { refresh: true }),
      // Sub-second precision must sort by time, not as a string.
      token("newest", "2026-11-20T10:00:00.5Z"),
      token("newer", "2026-11-20T10:00:00Z"),
    ];
    expect(selectPersonalAccessTokens(tokens).map((t) => t.id)).toEqual([
      "newest",
      "newer",
      "old",
    ]);
  });
});

describe("isExpiringSoon", () => {
  it.each([
    [undefined, false],
    ["2026-11-30T00:00:00Z", true],
    ["2026-12-02T14:30:15Z", true],
    ["2026-12-03T00:00:00Z", false],
  ])("treats %s as expiring soon: %s", (expiresOn, expected) => {
    expect(isExpiringSoon(expiresOn, now)).toBe(expected);
  });
});

describe("getTokenLastUsedOn", () => {
  it("treats a used_on equal to created_on as never used", () => {
    const createdOn = "2026-11-24T21:01:41.204662Z";
    expect(getTokenLastUsedOn({ createdOn, usedOn: createdOn })).toBe(
      undefined,
    );
    expect(
      getTokenLastUsedOn({ createdOn, usedOn: "2026-11-25T08:00:00Z" }),
    ).toBe("2026-11-25T08:00:00Z");
  });
});
