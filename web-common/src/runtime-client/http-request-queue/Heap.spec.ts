import { Heap } from "@rilldata/web-common/runtime-client/http-request-queue/Heap";
import { describe, it, expect } from "vitest";

type HeapTestItem = {
  name: string;
  order: string;
  value: number;

  index?: number;
};

describe("Heap", () => {
  it("should maintain add order", () => {
    const heap = new Heap<HeapTestItem>(
      (a, b) => a.value - b.value,
      (a) => a.name
    );

    heap.push({ name: "i0", order: "o0", value: 5 });
    heap.push({ name: "i1", order: "o1", value: 2 });
    heap.push({ name: "i2", order: "o2", value: 3 });
    heap.push({ name: "i3", order: "o3", value: 2 });
    heap.push({ name: "i3", order: "o4", value: 2 });
    heap.push({ name: "i4", order: "o5", value: 1 });
    heap.push({ name: "i0", order: "o6", value: 5 });

    const order = new Array<string>();
    while (!heap.empty()) {
      const item = heap.pop();
      order.push(item.order);
    }

    expect(order).toEqual(["o0", "o6", "o2", "o1", "o3", "o4", "o5"]);
  });
});
