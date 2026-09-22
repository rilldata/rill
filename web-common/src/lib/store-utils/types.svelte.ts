import { debounce } from "@rilldata/web-common/lib/create-debouncer.ts";

export interface RuneStore<Val, DefaultVal = Val> {
  value: Val | DefaultVal;
  getter: () => Val | DefaultVal;
  setter: (newValue: Val) => void;
}

export class InMemoryRuneStore<Val, DefaultVal = Val>
  implements RuneStore<Val, DefaultVal>
{
  public value: Val | DefaultVal;

  public constructor(defaultValue: DefaultVal) {
    this.value = $state(defaultValue);
  }

  public getter = () => this.value;

  public setter = (newValue: Val) => {
    this.value = newValue;
  };
}

/**
 * A convenience class to manage array of values stored in configured location.
 * Supported places are in-memory, url param or local storage (in a future PR)
 * Contains a method to toggle values.
 */
export class ArrayRuneStore<Val> implements RuneStore<Val[]> {
  public value: Val[];
  public getter: () => Val[];
  public setter: (newValue: Val[]) => void;

  public constructor(store: RuneStore<Val[]>) {
    this.value = $derived(store.value);
    this.getter = store.getter;
    this.setter = store.setter;
  }

  public toggle = (value: Val) => {
    const newValues = this.value.includes(value)
      ? this.value.filter((v) => v !== value)
      : [...this.value, value];
    this.setter(newValues);
  };

  public delete = (value: Val) => {
    const newValues = this.value.filter((v) => v !== value);
    this.setter(newValues);
  };
}

export class DebouncedRuneStore<Val, DefaultVal = Val>
  implements RuneStore<Val, DefaultVal>
{
  public value: Val;
  public getter: () => Val;
  public setter: (newValue: Val) => void;

  public constructor(
    private readonly store: RuneStore<Val>,
    timeout: number,
  ) {
    this.value = $derived(store.value);
    this.getter = store.getter;
    this.setter = debounce(store.setter, timeout);
  }

  public immediateSetter(newValue: Val) {
    (this.setter as ReturnType<typeof debounce>).cancel();
    this.store.setter(newValue);
  }
}
