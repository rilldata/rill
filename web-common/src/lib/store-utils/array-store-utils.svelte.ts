export interface ArrayStore<T> {
  value: T[];
  toggle: (value: T) => void;
}

export class InMemoryArrayStore<T> implements ArrayStore<T> {
  value: T[];

  public constructor(defaultValue: T[] = []) {
    this.value = $state(defaultValue);
  }

  public toggle = (value: T) => {
    const newValues = this.value.includes(value)
      ? this.value.filter((v) => v !== value)
      : [...this.value, value];
    this.value = newValues;
  };
}
