import {
  ArrayRuneStore,
  type RuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

export class SvelteLocalStorage<Val, DefaultVal>
  implements RuneStore<Val, DefaultVal>
{
  public value: Val | DefaultVal;

  // Cache of stores so that different components can share instance without prop drilling.
  // Since there is no event when localStorage is updated we need to ensure instances are shared.
  private static stores = new Map<string, SvelteLocalStorage<any, any>>();

  private constructor(
    private readonly key: string,
    private readonly serializer: (value: Val) => string | null,
    private readonly deserializer: (value: string | null) => Val | DefaultVal,
    defaultVal: Val | DefaultVal,
  ) {
    let initValue = defaultVal;
    try {
      const existingValue = localStorage.getItem(key);
      if (existingValue) {
        initValue = deserializer(existingValue);
      }
    } catch {
      // no-op
    }

    this.value = $state(initValue);
  }

  public static getInstance<Val, DefaultVal>(
    key: string,
    serializer: (value: Val) => string | null,
    deserializer: (value: string | null) => Val | DefaultVal,
    defaultVal: Val | DefaultVal,
  ) {
    if (this.stores.has(key))
      return this.stores.get(key) as SvelteLocalStorage<Val, DefaultVal>;
    const store = new SvelteLocalStorage(
      key,
      serializer,
      deserializer,
      defaultVal,
    );
    this.stores.set(key, store);
    return store;
  }

  /**
   * Drops the cache so that the next `getInstance` re-reads localStorage.
   * Instances already handed out keep the store they were given.
   */
  public static clearInstanceCache() {
    this.stores.clear();
  }

  /** For values that round-trip through JSON, such as objects and records. */
  public static createJsonStore<Val>(key: string, defaultVal: Val) {
    return SvelteLocalStorage.getInstance<Val, Val>(
      key,
      (value: Val) => JSON.stringify(value),
      (value) => (value ? (JSON.parse(value) as Val) : defaultVal),
      defaultVal,
    );
  }

  public static createStringArrayStore(key: string) {
    return new ArrayRuneStore<string>(
      SvelteLocalStorage.getInstance(
        key,
        (value: string[]) => (value.length ? JSON.stringify(value) : null),
        (value) => (value ? JSON.parse(value) : []),
        [],
      ),
    );
  }

  public getter = () => {
    return this.value;
  };

  public setter = (newValue: Val) => {
    const newStoreValue = this.serializer(newValue);
    try {
      if (newStoreValue) localStorage.setItem(this.key, newStoreValue);
      else localStorage.removeItem(this.key);
    } catch {
      // no-op
    }
    this.value = newValue;
  };
}
