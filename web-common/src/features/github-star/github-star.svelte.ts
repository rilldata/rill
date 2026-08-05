export const GITHUB_STAR_URL = "https://github.com/rilldata/rill";

const STORAGE_KEY = "rill:github-star";

const DAY_MS = 24 * 60 * 60 * 1000;

/**
 * Mute applied after the Nth soft dismiss.
 * Running off the end of the schedule is terminal: three ignores is a no.
 */
const MUTE_SCHEDULE_DAYS = [30, 90];

export type GithubStarStatus = "unarmed" | "armed" | "done";

export interface GithubStarState {
  status: GithubStarStatus;
  dismissCount: number;
  /** While in the future, an armed nudge stays hidden. */
  mutedUntil?: number;
}

const INITIAL_STATE: GithubStarState = { status: "unarmed", dismissCount: 0 };

/**
 * Tracks whether to nudge the user to star Rill on GitHub.
 * The nudge is armed by a dashboard render.
 * Dismissal escalates rather than muting for a fixed period. After three soft dismissals, the nudge is retired.
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
   * Recomputed whenever the state changes; mutes last 30+ days, so re-reading on
   * each app load is ample granularity for expiry.
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
   * Any other exit: X, click-outside, timeout, navigating away. These are
   * indistinguishable from the client, and not counting silent ignores would let
   * engaged users be re-asked every 30 days with no backoff.
   */
  public recordSoftDismiss() {
    if (!isVisible(this.currentState)) return;

    const dismissCount = this.currentState.dismissCount + 1;
    const muteDays = MUTE_SCHEDULE_DAYS[dismissCount - 1];
    if (muteDays === undefined) {
      this.commit({ ...this.currentState, status: "done", dismissCount });
      return;
    }
    this.commit({
      ...this.currentState,
      dismissCount,
      mutedUntil: Date.now() + muteDays * DAY_MS,
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
    // The worst case is one extra ask, which the backoff then bounds.
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
