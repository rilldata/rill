export class Throttler {
  private timer: ReturnType<typeof setTimeout> | undefined;
  private callback: () => void | Promise<void>;

  public constructor(
    private readonly timeout: number,
    private readonly shortTimeout: number,
  ) {}

  public throttle(
    callback: () => void | Promise<void>,
    useShortTimeout = false,
  ) {
    this.callback = callback;
    if (this.timer) return;

    this.timer = setTimeout(
      () => {
        this.timer = undefined;
        this.callback()?.catch(console.error);
      },
      useShortTimeout ? this.shortTimeout : this.timeout,
    );
  }

  public isThrottling() {
    return this.timer !== undefined;
  }

  public cancel() {
    clearTimeout(this.timer);
    this.timer = undefined;
  }
}
