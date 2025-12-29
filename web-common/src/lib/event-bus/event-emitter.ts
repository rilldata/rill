type VoidType = void | Promise<void>;
type Listener<Events extends Record<string, any>, E extends keyof Events> = (
  arg: Events[E],
) => VoidType;

export class EventEmitter<Events extends Record<string, any>> {
  private readonly listeners = new Map<
    keyof Events,
    Map<string, Listener<Events, keyof Events>>
  >();

  public on<E extends keyof Events>(event: E, listener: Listener<Events, E>) {
    const key = generateUUID();
    const eventMap = this.listeners.get(event);

    if (!eventMap) {
      this.listeners.set(event, new Map([[key, listener]]));
    } else {
      eventMap.set(key, listener);
    }

    const unsubscribe = () => this.listeners.get(event)?.delete(key);

    return unsubscribe;
  }

  once<E extends keyof Events>(event: E, callback: Listener<Events, E>) {
    const unsubscribe = this.on(event, ((payload) => {
      callback(payload);
      unsubscribe();
    }) as any);

    return unsubscribe;
  }

  public emit<E extends keyof Events>(event: E, arg: Events[E]) {
    const listeners = this.listeners.get(event);

    listeners?.forEach((listener) => {
      listener(arg);
    });
  }

  public clearListeners() {
    this.listeners.clear();
  }
}

function generateUUID(): string {
  // Generate random numbers for the UUID
  const randomNumbers: number[] = new Array(16)
    .fill(0)
    .map(() => Math.floor(Math.random() * 256));

  // Set the version and variant bits
  randomNumbers[6] = (randomNumbers[6] & 0x0f) | 0x40; // Version 4
  randomNumbers[8] = (randomNumbers[8] & 0x3f) | 0x80; // Variant 10

  // Convert to hexadecimal and format as a UUID
  const hexDigits: string = randomNumbers
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
  return `${hexDigits.slice(0, 8)}-${hexDigits.slice(8, 12)}-${hexDigits.slice(12, 16)}-${hexDigits.slice(16, 20)}-${hexDigits.slice(20, 32)}`;
}
