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
  private currentState = $state<GithubStarState>(INITIAL_STATE);

  public constructor() {
    this.currentState = readState();
  }

  public get state() {
    return this.currentState;
  }

  /**
   * Whether the nudge should be shown right now.
   * Recomputed whenever the state changes; re-reading on each app load is ample
   * granularity for the one-day mute to expire.
   */
  public get visible() {
    return isVisible(this.currentState);
  }

  /** Called when the user reaches a payoff moment: a dashboard rendered. */
  public armPayoff() {
    if (this.currentState.status !== "unarmed") return;
    this.commit({ ...this.currentState, status: "armed" });
  }

  /** The user starred. Terminal. */
  public recordStar() {
    this.commit({ ...this.currentState, status: "done" });
  }

  /** The user clicked "Don't show again". Terminal. */
  public recordOptOut() {
    this.commit({ ...this.currentState, status: "done" });
  }

  /**
   * A non-terminal exit, such as Escape, X, or click-outside, mutes the nudge
   * until the following day.
   */
  public recordSoftDismiss() {
    if (!isVisible(this.currentState)) return;

    this.commit({
      ...this.currentState,
      mutedUntil: Date.now() + DAY_MS,
    });
  }

  private commit(next: GithubStarState) {
    this.currentState = next;
    writeState(next);
  }
}

export function isVisible(state: GithubStarState, now = Date.now()) {
  if (state.status !== "armed") return false;
  return !state.mutedUntil || state.mutedUntil <= now;
}

function readState(): GithubStarState {
  try {
    // Accessing localStorage can throw rather than return null: a sandboxed or
    // storage-partitioned iframe raises a SecurityError, and it is absent on the server.
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return INITIAL_STATE;
    const parsed = JSON.parse(raw) as GithubStarState;
    if (!parsed || typeof parsed.status !== "string") return INITIAL_STATE;
    return { ...INITIAL_STATE, ...parsed };
  } catch {
    // Corrupt or unreadable state degrades to "unarmed" rather than throwing.
    // The worst case is one extra ask.
    return INITIAL_STATE;
  }
}

function writeState(state: GithubStarState) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch {
    // no-op: nudge degrades to in-memory only
  }
}

export const githubStarNudge = new GithubStarNudge();
