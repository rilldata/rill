import { v4 as uuidv4 } from "uuid";

type VoidType = void | Promise<void>;
type Listener<
  Events extends Record<string, unknown>,
  E extends keyof Events,
> = Events[E] extends void ? () => VoidType : (arg: Events[E]) => VoidType;
type Args<
  Events extends Record<string, unknown>,
  E extends keyof Events,
> = Events[E] extends void ? [] : [Events[E]];

export class EventEmitter<Events extends Record<string, unknown>> {
  private readonly listeners = new Map<
    keyof Events,
    Map<string, Listener<Events, keyof Events>>
  >();

  public on<E extends keyof Events>(event: E, listener: Listener<Events, E>) {
    const key = uuidv4();
    const eventMap = this.listeners.get(event);

    if (!eventMap) {
      this.listeners.set(event, new Map([[key, listener]]));
    } else {
      eventMap.set(key, listener);
    }

    const unsubscribe = () => this.listeners.get(event)?.delete(key);

    return unsubscribe;
  }

  once<E extends keyof Events>(event: E, listener: Listener<Events, E>) {
    const unsubscribe = this.on(event, ((...args: Args<Events, E>) => {
      (listener as any)(...args);
      unsubscribe();
    }) as Listener<Events, E>);

    return unsubscribe;
  }

  public emit<E extends keyof Events>(event: E, ...args: Args<Events, E>) {
    const listeners = this.listeners.get(event);

    listeners?.forEach((listener) => {
      (listener as any)(...args);
    });
  }

  public clearListeners() {
    this.listeners.clear();
  }
}
