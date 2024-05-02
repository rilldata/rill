export class ExponentialBackoffTracker {
  private curRetries = 0;
  private currentDelay: number;

  public constructor(
    private readonly maxRetries: number,

    private readonly initialDelay: number,
  ) {
    this.currentDelay = initialDelay;
  }

  public static createBasicTracker() {
    return new ExponentialBackoffTracker(5, 1000);
  }

  public try = async (fn: () => Promise<void> | void) => {
    try {
      await fn();

      this.curRetries = 0;
      this.currentDelay = this.initialDelay;
    } catch (e) {
      if (this.curRetries >= this.maxRetries) {
        throw e;
      }

      this.currentDelay = this.initialDelay * 2 ** this.curRetries;
      await new Promise((resolve) => setTimeout(resolve, this.currentDelay));

      this.curRetries++;
      return this.try(fn);
    }
  };
}
