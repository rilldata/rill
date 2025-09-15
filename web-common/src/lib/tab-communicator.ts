export class TabCommunicator<T> {
  private readonly channel: BroadcastChannel;
  private readonly listeners: Map<string, ((data: T) => void)[]> = new Map();

  constructor(channelName: string) {
    this.channel = new BroadcastChannel(channelName);
    this.channel.addEventListener("message", (event) =>
      this.handleMessage(event),
    );
  }

  public send(type: string, data: T) {
    this.channel.postMessage({ type, data });
  }

  public on(type: string, callback: (data: T) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, []);
    }
    this.listeners.get(type)!.push(callback);
  }

  public off(type: string, callback: (data: T) => void) {
    const callbacks = this.listeners.get(type);
    if (!callbacks) return;
    const index = callbacks.indexOf(callback);
    if (index === -1) return;
    callbacks.splice(index, 1);
  }

  public close() {
    this.channel.close();
  }

  private handleMessage(event: MessageEvent) {
    const { type, data } = event.data;
    const callbacks = this.listeners.get(type) || [];
    callbacks.forEach((callback) => callback(data));
  }
}
