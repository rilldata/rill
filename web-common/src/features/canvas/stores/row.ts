import { writable, get } from "svelte/store";
import { Grid } from "./grid";
import { COLUMN_COUNT } from "../layout-util";

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

    let newLength = 0;

    this.items.update((arr) => {
      const copy = [...arr];
      copy.splice(index, 1);
      newLength = copy.length;
      return copy;
    });

    this.widths.update(() => {
      return Array.from({ length: newLength }, () => COLUMN_COUNT / newLength);
    });

    if (newLength === 0) {
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
