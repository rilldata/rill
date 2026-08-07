import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GithubStarNudge } from "./github-star.svelte";

const DAY_MS = 24 * 60 * 60 * 1000;
const START = new Date("2026-08-04T12:00:00Z").getTime();

/** Sets the clock to `days` after the start of the test. */
function atDay(days: number) {
  vi.setSystemTime(START + days * DAY_MS);
}

/** A fresh nudge that re-reads localStorage, standing in for an app reload. */
function reload() {
  SvelteLocalStorage.clearInstanceCache();
  return new GithubStarNudge();
}

describe("GithubStarNudge", () => {
  beforeEach(() => {
    localStorage.clear();
    SvelteLocalStorage.clearInstanceCache();
    vi.useFakeTimers();
    vi.setSystemTime(START);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("stays hidden until a payoff arms it", () => {
    const nudge = new GithubStarNudge();
    expect(nudge.visible).toBe(false);

    nudge.armPayoff();
    expect(nudge.visible).toBe(true);
  });

  it("persists the armed state across reloads", () => {
    new GithubStarNudge().armPayoff();
    expect(reload().visible).toBe(true);
  });

  it("treats a star click as terminal", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordStar();
    expect(nudge.visible).toBe(false);

    // Neither a later payoff nor the passage of time may resurrect it.
    nudge.armPayoff();
    expect(nudge.visible).toBe(false);
    atDay(365);
    expect(reload().visible).toBe(false);
  });

  it("treats an explicit opt-out as terminal", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordOptOut();

    atDay(365);
    expect(reload().visible).toBe(false);
  });

  it("mutes a soft dismissal until the following day", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordSoftDismiss();

    expect(nudge.state.status).toBe("armed");
    expect(nudge.visible).toBe(false);
    atDay(0.99);
    expect(reload().visible).toBe(false);
    atDay(1);
    expect(reload().visible).toBe(true);
  });

  it("continues asking daily until the user stars or opts out", () => {
    new GithubStarNudge().armPayoff();

    let asks = 0;
    for (let day = 0; day < 400; day++) {
      atDay(day);
      const nudge = reload();
      if (nudge.visible) {
        asks++;
        nudge.recordSoftDismiss();
      }
    }

    expect(asks).toBe(400);
    expect(reload().state.status).toBe("armed");
  });

  it("ignores a soft dismissal while muted", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordSoftDismiss();
    const mutedUntil = nudge.state.mutedUntil;

    // A dismissal that could not have been seen must not extend the mute.
    nudge.recordSoftDismiss();
    expect(nudge.state.mutedUntil).toBe(mutedUntil);
  });

  it("ignores a soft dismissal that was never armed", () => {
    const nudge = new GithubStarNudge();
    nudge.recordSoftDismiss();

    expect(nudge.state.status).toBe("unarmed");
  });

  it("does not re-arm on a second payoff once muted", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordSoftDismiss();

    nudge.armPayoff();
    expect(nudge.visible).toBe(false);
  });

  it("treats corrupt stored state as unarmed", () => {
    localStorage.setItem("rill:github-star", "{not json");
    expect(reload().visible).toBe(false);

    localStorage.setItem("rill:github-star", JSON.stringify({ nonsense: 1 }));
    expect(reload().visible).toBe(false);
  });

  it("degrades to in-memory when localStorage throws", () => {
    const setItem = vi
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new DOMException("denied", "SecurityError");
      });
    const getItem = vi
      .spyOn(Storage.prototype, "getItem")
      .mockImplementation(() => {
        throw new DOMException("denied", "SecurityError");
      });

    const nudge = new GithubStarNudge();
    expect(() => nudge.armPayoff()).not.toThrow();
    expect(nudge.visible).toBe(true);

    setItem.mockRestore();
    getItem.mockRestore();
  });
});
