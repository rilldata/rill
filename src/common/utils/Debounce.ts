export class Debounce {
  private debounceMap = new Map<string, NodeJS.Timer>();

  public debounce(id: string, callback: () => any, time: number) {
    if (this.debounceMap.has(id)) return;

    this.debounceMap.set(
      id,
      setTimeout(() => {
        this.debounceMap.delete(id);
        callback();
      }, time)
    );
  }

  public clear(id: string) {
    if (this.debounceMap.has(id)) {
      clearTimeout(this.debounceMap.get(id));
      this.debounceMap.delete(id);
    }
  }
}
