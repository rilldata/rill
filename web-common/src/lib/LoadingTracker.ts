import { writable } from "svelte/store";

/**
 * Tracks loading state with support for loading for short and long status.
 * Pass in shortLoadDelay and longLoadDelay to control the behavior.
 */
export class LoadingTracker {
  public readonly loadingForShortTime = writable(false);
  public readonly loadingForLongTime = writable(false);

  private loadingForShortTimeout: ReturnType<typeof setTimeout> | undefined;
  private loadingForLongTimeout: ReturnType<typeof setTimeout> | undefined;

  public constructor(
    private readonly shortLoadDelay: number,
    private readonly longLoadDelay: number,
  ) {}

  public updateLoading(loading: boolean) {
    if (loading) {
      this.startedLoading();
    } else {
      this.endedLoading();
    }
  }

  private startedLoading() {
    if (!this.loadingForShortTimeout) {
      this.loadingForShortTimeout = setTimeout(() => {
        this.loadingForShortTime.set(true);
        this.loadingForShortTimeout = undefined;
      }, this.shortLoadDelay);
    }

    if (!this.loadingForLongTimeout) {
      this.loadingForLongTimeout = setTimeout(() => {
        this.loadingForLongTime.set(true);
        this.loadingForLongTimeout = undefined;
      }, this.longLoadDelay);
    }
  }

  private endedLoading() {
    this.loadingForShortTime.set(false);
    this.loadingForLongTime.set(false);

    if (this.loadingForShortTimeout) clearTimeout(this.loadingForShortTimeout);
    this.loadingForShortTimeout = undefined;
    if (this.loadingForLongTimeout) clearTimeout(this.loadingForLongTimeout);
    this.loadingForLongTimeout = undefined;
  }
}
