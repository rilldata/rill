import { SvelteLocalStorage } from "@rilldata/web-common/lib/store-utils/svelte-local-storage.svelte.ts";
import type { RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

export const GITHUB_STAR_URL = "https://github.com/rilldata/rill";

const STORAGE_KEY = "rill:github-star";

const DAY_MS = 24 * 60 * 60 * 1000;

export type GithubStarStatus = "unarmed" | "armed" | "done";

export interface GithubStarState {
  status: GithubStarStatus;
  /** While in the future, an armed nudge stays hidden. */
  mutedUntil?: number;
}

const INITIAL_STATE: GithubStarState = { status: "unarmed" };

/**
 * Tracks whether to nudge the user to star Rill on GitHub.
 * The nudge is armed by a dashboard render.
 * A soft dismissal mutes it for one day; only starring or opting out retires it.
 */
export class GithubStarNudge {
  public constructor(
    private readonly store: RuneStore<GithubStarState> = SvelteLocalStorage.createJsonStore<GithubStarState>(
      STORAGE_KEY,
      INITIAL_STATE,
    ),
  ) {}

  public get state() {
    return this.store.value;
  }

  /**
   * Whether the nudge should be shown right now.
   * Recomputed whenever the state changes; re-reading on each app load is ample
   * granularity for the one-day mute to expire.
   */
  public get visible() {
    const { status, mutedUntil } = this.store.value;
    if (status !== "armed") return false;
    return !mutedUntil || mutedUntil <= Date.now();
  }

  /** Called when the user reaches a payoff moment: a dashboard rendered. */
  public armPayoff() {
    const { status } = this.store.value;
    if (status === "armed" || status === "done") return;
    this.store.setter({ ...this.store.value, status: "armed" });
  }

  /** The user starred. Terminal. */
  public recordStar() {
    this.store.setter({ ...this.store.value, status: "done" });
  }

  /** The user clicked "Don't show again". Terminal. */
  public recordOptOut() {
    this.store.setter({ ...this.store.value, status: "done" });
  }

  /**
   * A non-terminal exit, such as Escape, X, or click-outside, mutes the nudge
   * until the following day.
   */
  public recordSoftDismiss() {
    if (!this.visible) return;

    this.store.setter({
      ...this.store.value,
      mutedUntil: Date.now() + DAY_MS,
    });
  }
}

export const githubStarNudge = new GithubStarNudge();
