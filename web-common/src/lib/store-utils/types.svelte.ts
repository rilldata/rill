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

  getter(): Val | DefaultVal {
    return this.value;
  }

  setter(newValue: Val): void {
    this.value = newValue;
  }
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
    const newTags = this.value.includes(value)
      ? this.value.filter((v) => v !== value)
      : [...this.value, value];
    this.setter(newTags);
  };
}
