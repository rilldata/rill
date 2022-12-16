export class Debounce {
  private debounceMap = new Map<string, NodeJS.Timer>();
  private callbackMap = new Map<string, () => void | Promise<void>>();

  public debounce(
    id: string,
    callback: () => void | Promise<void>,
    time: number
  ) {
    this.callbackMap.set(id, callback);
    if (this.debounceMap.has(id)) return;

    this.debounceMap.set(
      id,
      setTimeout(() => {
        this.debounceMap.delete(id);
        this.callbackMap.get(id)();
      }, time)
    );
  }

  public clear(id: string) {
    if (this.debounceMap.has(id)) {
      clearTimeout(this.debounceMap.get(id));
      this.debounceMap.delete(id);
      this.callbackMap.delete(id);
    }
  }
}
