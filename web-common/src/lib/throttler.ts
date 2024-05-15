export class Throttler {
  private timer: ReturnType<typeof setTimeout> | undefined;
  private callback: () => void | Promise<void>;

  public constructor(private readonly timeout: number) {}

  public throttle(callback: () => void | Promise<void>) {
    this.cancel();
    this.callback = callback;

    this.timer = setTimeout(() => {
      this.timer = undefined;
      this.callback();
    }, this.timeout);
  }

  public isThrottling() {
    return this.timer !== undefined;
  }

  public cancel() {
    if (!this.timer) return;

    clearTimeout(this.timer);
    this.timer = undefined;
  }
}

export class ThrottlerMap extends Map<
  string,
  [ReturnType<typeof setTimeout>, () => void | Promise<void>]
> {
  public constructor(private readonly timeout: number) {
    super();
  }

  public throttle(id: string, callback: () => void | Promise<void>) {
    this.cancel(id);

    const timer = setTimeout(() => {
      const entry = this.get(id);
      if (!entry) return;

      clearTimeout(entry[0]);
      void entry[1]();
      this.delete(id);
    }, this.timeout);

    this.set(id, [timer, callback]);
  }

  public cancel(id: string): void {
    const entry = this.get(id);
    if (!entry) return;
    clearTimeout(entry[0]);
    this.delete(id);
  }
}
