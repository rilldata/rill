/**
 * Keeps the items pushed in the last `windowMs` milliseconds, at most `maxItems` of the most recent.
 * Older items are dropped on every push and read, so the buffer never grows past `maxItems`.
 */
export class TimeBoundedBuffer<T> {
  private readonly entries: Array<{ at: number; item: T }> = [];

  public constructor(
    private readonly windowMs: number,
    private readonly maxItems: number,
  ) {}

  public push(item: T): void {
    const at = performance.now();
    this.entries.push({ at, item });
    this.prune(at);
  }

  /** The items pushed within the window with the performance.now() time of each push, most recent first. */
  public recent(): Array<{ at: number; item: T }> {
    this.prune(performance.now());
    return this.entries.slice().reverse();
  }

  private prune(now: number) {
    const firstInWindow = this.entries.findIndex(
      (entry) => now - entry.at <= this.windowMs,
    );
    const expired = firstInWindow === -1 ? this.entries.length : firstInWindow;
    this.entries.splice(
      0,
      Math.max(expired, this.entries.length - this.maxItems),
    );
  }
}
