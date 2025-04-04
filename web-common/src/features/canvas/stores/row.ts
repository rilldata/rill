import { writable, get } from "svelte/store";
import { Grid } from "./grid";

export class Row {
  items = writable<string[]>([]);
  widths = writable<number[]>([]);
  height = writable<number>(100);

  constructor(private grid: Grid) {}

  add(value: string, index?: number) {
    this.items.update((arr) => {
      const copy = [...arr];
      if (index === undefined) {
        copy.push(value);
      } else {
        copy.splice(index - 1, 0, value);
      }
      return copy;
    });
  }

  remove(index: number) {
    if (index < 1) return;
    this.items.update((arr) => {
      const copy = [...arr];
      copy.splice(index - 1, 1);
      return copy;
    });

    if (this.isEmpty()) {
      this.grid.removeRow(this);
    }
  }

  moveWithin(fromIndex: number, toIndex: number) {
    if (fromIndex === toIndex) return;
    this.items.update((arr) => {
      const copy = [...arr];
      const value = copy.splice(fromIndex - 1, 1)[0];
      copy.splice(toIndex - 1, 0, value);
      return copy;
    });
  }

  getValue(index: number): string | null {
    return get(this.items)[index] ?? null;
  }

  isEmpty(): boolean {
    return get(this.items).length === 0;
  }

  snapshot(): string[] {
    return get(this.items);
  }
}
