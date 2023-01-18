/**
 * Makes sure only one instance of the callback is running at a time
 */
export class Throttler<
  Func extends (...args: Array<unknown>) => Promise<void>
> {
  private throttlerMap = new Map<string, Promise<void>>();
  private argumentsMap = new Map<string, Parameters<Func>>();

  public constructor(private readonly callback: Func) {}

  public throttle(id: string, args: Parameters<Func>) {
    this.argumentsMap.set(id, args);
    if (this.throttlerMap.has(id)) return;

    const promise = this.callback(...this.argumentsMap.get(id));
    this.throttlerMap.set(id, promise);
    promise.then(
      () => {
        this.clear(id);
      },
      () => {
        this.clear(id);
      }
    );
  }

  private clear(id: string) {
    this.throttlerMap.delete(id);
  }
}
