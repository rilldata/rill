import { afterEach, describe, expect, it, vi } from "vitest";
import { TimeBoundedBuffer } from "./time-bounded-buffer";

describe("TimeBoundedBuffer", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  function setNow(now: number) {
    vi.spyOn(performance, "now").mockReturnValue(now);
  }

  it("returns the items within the window, most recent first", () => {
    const buffer = new TimeBoundedBuffer<string>(100, 10);
    setNow(1000);
    buffer.push("first");
    setNow(1050);
    buffer.push("second");

    expect(buffer.recent()).toEqual([
      { at: 1050, item: "second" },
      { at: 1000, item: "first" },
    ]);

    setNow(1120);
    expect(buffer.recent()).toEqual([{ at: 1050, item: "second" }]);

    setNow(1200);
    expect(buffer.recent()).toEqual([]);
  });

  it("keeps at most the most recent items when more arrive within the window", () => {
    const buffer = new TimeBoundedBuffer<number>(100, 3);
    setNow(1000);
    for (let i = 0; i < 5; i++) buffer.push(i);

    expect(buffer.recent().map((entry) => entry.item)).toEqual([4, 3, 2]);
  });
});
