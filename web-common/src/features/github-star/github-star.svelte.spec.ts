import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GithubStarNudge } from "./github-star.svelte";

const DAY_MS = 24 * 60 * 60 * 1000;
const START = new Date("2026-08-04T12:00:00Z").getTime();

/** Sets the clock to `days` after the start of the test. */
function atDay(days: number) {
  vi.setSystemTime(START + days * DAY_MS);
}

describe("GithubStarNudge", () => {
  beforeEach(() => {
    localStorage.clear();
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
    expect(new GithubStarNudge().visible).toBe(true);
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
    expect(new GithubStarNudge().visible).toBe(false);
  });

  it("treats an explicit opt-out as terminal", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordOptOut();

    atDay(365);
    expect(new GithubStarNudge().visible).toBe(false);
  });

  it("escalates soft dismissals to 30 days, then 90 days, then permanently", () => {
    const first = new GithubStarNudge();
    first.armPayoff();

    first.recordSoftDismiss();
    expect(first.visible).toBe(false);
    atDay(29);
    expect(new GithubStarNudge().visible).toBe(false);
    atDay(31);
    expect(new GithubStarNudge().visible).toBe(true);

    const second = new GithubStarNudge();
    second.recordSoftDismiss();
    atDay(31 + 89);
    expect(new GithubStarNudge().visible).toBe(false);
    atDay(31 + 91);
    expect(new GithubStarNudge().visible).toBe(true);

    const third = new GithubStarNudge();
    third.recordSoftDismiss();
    expect(third.state.status).toBe("done");
    atDay(3650);
    expect(new GithubStarNudge().visible).toBe(false);
  });

  it("caps an install at three asks", () => {
    new GithubStarNudge().armPayoff();

    let asks = 0;
    for (let day = 0; day < 400; day++) {
      atDay(day);
      const nudge = new GithubStarNudge();
      if (nudge.visible) {
        asks++;
        nudge.recordSoftDismiss();
      }
    }

    expect(asks).toBe(3);
  });

  it("ignores a soft dismissal while muted", () => {
    const nudge = new GithubStarNudge();
    nudge.armPayoff();
    nudge.recordSoftDismiss();
    expect(nudge.state.dismissCount).toBe(1);

    // A dismissal that could not have been seen must not advance the backoff.
    nudge.recordSoftDismiss();
    expect(nudge.state.dismissCount).toBe(1);
  });

  it("ignores a soft dismissal that was never armed", () => {
    const nudge = new GithubStarNudge();
    nudge.recordSoftDismiss();

    expect(nudge.state.dismissCount).toBe(0);
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
    expect(new GithubStarNudge().visible).toBe(false);

    localStorage.setItem("rill:github-star", JSON.stringify({ nonsense: 1 }));
    expect(new GithubStarNudge().visible).toBe(false);
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
