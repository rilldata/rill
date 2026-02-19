import { describe, it, expect, vi, afterEach } from "vitest";
import { getRelativeTime, timeAgo } from "./relative-time";

describe("getRelativeTime", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns empty string for empty input", () => {
    expect(getRelativeTime("")).toBe("");
  });

  it("returns empty string for invalid date", () => {
    expect(getRelativeTime("not-a-date")).toBe("");
  });

  it('returns "now" for dates less than 1 minute ago', () => {
    const now = new Date();
    expect(getRelativeTime(now.toISOString())).toBe("now");
  });

  it("returns minutes ago for dates less than 1 hour ago", () => {
    vi.useFakeTimers();
    const now = new Date("2026-01-15T12:00:00Z");
    vi.setSystemTime(now);

    const fiveMinutesAgo = new Date("2026-01-15T11:55:00Z");
    expect(getRelativeTime(fiveMinutesAgo.toISOString())).toBe("5m ago");
  });

  it("returns hours ago for dates less than 24 hours ago", () => {
    vi.useFakeTimers();
    const now = new Date("2026-01-15T12:00:00Z");
    vi.setSystemTime(now);

    const threeHoursAgo = new Date("2026-01-15T09:00:00Z");
    expect(getRelativeTime(threeHoursAgo.toISOString())).toBe("3h ago");
  });

  it("returns empty string for dates more than 24 hours ago", () => {
    vi.useFakeTimers();
    const now = new Date("2026-01-15T12:00:00Z");
    vi.setSystemTime(now);

    const twoDaysAgo = new Date("2026-01-13T12:00:00Z");
    expect(getRelativeTime(twoDaysAgo.toISOString())).toBe("");
  });
});

describe("timeAgo", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "Just now" for dates less than 1 minute ago', () => {
    vi.useFakeTimers();
    const now = new Date("2026-01-15T12:00:00Z");
    vi.setSystemTime(now);

    expect(timeAgo(new Date("2026-01-15T11:59:45Z"))).toBe("Just now");
  });

  it("returns singular minute", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-15T12:01:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("1 minute ago");
  });

  it("returns plural minutes", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-15T12:05:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("5 minutes ago");
  });

  it("returns singular hour", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-15T13:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("1 hour ago");
  });

  it("returns plural hours", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-15T15:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("3 hours ago");
  });

  it("returns singular day", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-16T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("1 day ago");
  });

  it("returns plural days", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-18T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("3 days ago");
  });

  it("returns singular week", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-22T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("1 week ago");
  });

  it("returns plural weeks", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-02-05T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("3 weeks ago");
  });

  it("returns singular month", () => {
    vi.useFakeTimers();
    // 38 days = ~5.4 weeks (past the 5-week threshold) and ~1.3 months (rounds to 1)
    vi.setSystemTime(new Date("2026-02-22T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("1 month ago");
  });

  it("returns plural months", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-06-15T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("5 months ago");
  });

  it("returns singular year", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2027-01-15T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("1 year ago");
  });

  it("returns plural years", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2029-01-15T12:00:00Z"));

    expect(timeAgo(new Date("2026-01-15T12:00:00Z"))).toBe("3 years ago");
  });
});
