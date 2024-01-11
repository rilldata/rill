import { asyncWait } from "@rilldata/web-common/lib/waitUtils";

export class ExponentialBackoffTracker {
  private curTime: number;
  private curRetries = 0;
  private trackerPeriod: number;

  public constructor(
    private readonly retries: number,
    /**
     * Time period within which to trigger the tracker.
     * Any failure after this will be considered as an intermittent failure.
     */
    private readonly trackerTriggerPeriod: number,
    // time
    private readonly waitPeriod: number,
  ) {
    this.trackerPeriod = trackerTriggerPeriod;
  }

  public static createBasicTracker() {
    return new ExponentialBackoffTracker(5, 1000, 250);
  }

  public async failed(): Promise<boolean> {
    const lastTime = this.curTime;
    this.curTime = Date.now();

    // if failed after the tracker period, reset everything
    if (this.curTime - lastTime >= this.trackerPeriod) {
      this.reset();
      await asyncWait(this.waitPeriod);
      return true;
    }

    // if retry count has reached max return false to stop the connection
    if (this.curRetries === this.retries) return false;

    // increment retry and update the tracker periods.
    this.curRetries++;
    // A simple function to increase the tracking period.
    this.trackerPeriod = this.trackerTriggerPeriod * 2 ** this.curRetries;
    // multiply the retires to the wait period as well.
    await asyncWait(this.waitPeriod * 2 ** this.curRetries);
    return true;
  }

  private reset() {
    this.curRetries = 0;
    this.trackerPeriod = this.trackerTriggerPeriod;
  }
}
