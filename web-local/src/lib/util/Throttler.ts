export class Throttler {
  private throttleMap = new Map<string, NodeJS.Timer>();

  public throttle(id: string, callback: () => void, time: number) {
    if (this.throttleMap.has(id)) return;

    this.throttleMap.set(
      id,
      setTimeout(() => {
        this.throttleMap.delete(id);
      }, time)
    );
    // call the callback at the beginning of the timer.
    // but an entry in throttleMap makes it so that this is not called multiple times within `time`ms
    callback();
  }

  public clear(id: string) {
    if (!this.throttleMap.has(id)) return;
    clearTimeout(this.throttleMap.get(id));
    this.throttleMap.delete(id);
  }
}
