export class Heap<Item = any> {
  private readonly array: Array<Item>;
  private valueToIdxMap: Map<Item, number> = new Map();
  private readonly compareFunction: (a: Item, b: Item) => number;

  /**
   * @constructor
   * @param {Array} initArray
   * @param {Function} compareFunction Return value > 0 to have a above b in the heap.
   */
  constructor(
    initArray = [],
    compareFunction = function (a, b) {
      return a - b;
    }
  ) {
    this.array = initArray;
    this.compareFunction = compareFunction;

    this.array.forEach((e, i) => {
      this.valueToIdxMap.set(e, i);
    });
  }

  public empty() {
    return this.array.length === 0;
  }

  public peek(): Item {
    return this.array[0];
  }

  public push(value: Item) {
    this.valueToIdxMap.set(value, this.array.length);
    this.array.push(value);
    this.moveUp(this.array.length - 1);
  }

  public pop() {
    if (this.array.length > 0) {
      const value = this.array[0];
      this.valueToIdxMap.delete(value);
      if (this.array.length > 1) {
        this.array[0] = this.array.pop();
        this.valueToIdxMap.set(this.array[0], 0);
        this.moveDown(0);
      } else {
        this.array.pop();
      }
      return value;
    }
  }

  public delete(value: Item) {
    const idx = this.valueToIdxMap.get(value);
    if (idx >= 0) {
      this.valueToIdxMap.delete(value);
      if (idx < this.array.length - 1) {
        this.array[idx] = this.array.pop();
        this.valueToIdxMap.set(this.array[idx], idx);
        this.moveDown(idx);
      } else {
        this.array.pop();
      }
    }
  }

  // doesnt work on literals
  public updateItem(value: Item) {
    const idx = this.valueToIdxMap.get(value);
    if (!this.moveUp(idx)) {
      this.moveDown(idx);
    }
  }

  private moveUp(idx: number) {
    let movedUp = false;
    while (idx > 0) {
      const parentIdx = (idx - 1) >> 1;
      if (this.compareFunction(this.array[idx], this.array[parentIdx]) > 0) {
        this.swap(idx, parentIdx);
        idx = parentIdx;
        movedUp = true;
      } else {
        break;
      }
    }
    return movedUp;
  }

  private moveDown(idx: number) {
    let movedDown = false;
    while (idx < this.array.length) {
      let childIdx = 2 * idx + 1;

      if (childIdx >= this.array.length) {
        break;
      }

      if (
        childIdx + 1 < this.array.length &&
        this.compareFunction(this.array[childIdx + 1], this.array[childIdx]) > 0
      ) {
        childIdx++;
      }

      if (this.compareFunction(this.array[childIdx], this.array[idx]) > 0) {
        this.swap(idx, childIdx);
        idx = childIdx;
        movedDown = true;
      } else {
        break;
      }
    }
    return movedDown;
  }

  private swap(idx0: number, idx1: number) {
    const val0 = this.array[idx0];
    this.array[idx0] = this.array[idx1];
    this.array[idx1] = val0;
    this.valueToIdxMap.set(this.array[idx0], idx0);
    this.valueToIdxMap.set(this.array[idx1], idx1);
  }
}
