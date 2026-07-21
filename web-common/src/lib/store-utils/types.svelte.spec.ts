import { describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";
import {
  ArrayRuneStore,
  DebouncedRuneStore,
  InMemoryRuneStore,
} from "./types.svelte.ts";

describe("InMemoryRuneStore", () => {
  it("holds and updates the value", () => {
    const cleanup = $effect.root(() => {
      const store = new InMemoryRuneStore<string>("");
      expect(store.value).toBe("");
      expect(store.getter()).toBe("");

      store.setter("hello");
      expect(store.value).toBe("hello");
      expect(store.getter()).toBe("hello");
    });
    cleanup();
  });

  it("exposes bound getter/setter that can be detached", () => {
    const cleanup = $effect.root(() => {
      const store = new InMemoryRuneStore<string>("");
      const { getter, setter } = store;

      setter("detached");
      expect(getter()).toBe("detached");
      expect(store.value).toBe("detached");
    });
    cleanup();
  });
});

describe("ArrayRuneStore", () => {
  it("toggle adds a missing value and removes an existing one", () => {
    const cleanup = $effect.root(() => {
      const store = new ArrayRuneStore<string>(
        new InMemoryRuneStore<string[]>([]),
      );
      flushSync();

      store.toggle("a");
      flushSync();
      expect(store.value).toEqual(["a"]);

      store.toggle("b");
      flushSync();
      expect(store.value).toEqual(["a", "b"]);

      store.toggle("a");
      flushSync();
      expect(store.value).toEqual(["b"]);
    });
    cleanup();
  });

  it("delete removes the value and is a no-op for absent values", () => {
    const cleanup = $effect.root(() => {
      const store = new ArrayRuneStore<string>(
        new InMemoryRuneStore<string[]>(["a", "b", "c"]),
      );
      flushSync();

      store.delete("b");
      flushSync();
      expect(store.value).toEqual(["a", "c"]);

      store.delete("missing");
      flushSync();
      expect(store.value).toEqual(["a", "c"]);
    });
    cleanup();
  });

  it("reflects updates written to the underlying store", () => {
    const cleanup = $effect.root(() => {
      const underlying = new InMemoryRuneStore<string[]>([]);
      const store = new ArrayRuneStore<string>(underlying);
      flushSync();

      store.setter(["x", "y"]);
      flushSync();
      expect(store.value).toEqual(["x", "y"]);
      expect(underlying.value).toEqual(["x", "y"]);
    });
    cleanup();
  });
});

describe("DebouncedRuneStore", () => {
  it("debounces writes to the underlying store", () => {
    vi.useFakeTimers();
    const cleanup = $effect.root(() => {
      const underlying = new InMemoryRuneStore<string>("");
      const store = new DebouncedRuneStore<string>(underlying, 500);
      flushSync();

      store.setter("a");
      flushSync();
      // The underlying store (and thus the derived value) is not updated yet.
      expect(underlying.value).toBe("");
      expect(store.value).toBe("");

      // A second write within the window resets the timer.
      vi.advanceTimersByTime(499);
      store.setter("ab");
      vi.advanceTimersByTime(499);
      flushSync();
      expect(underlying.value).toBe("");

      // Only the last write lands once the window elapses.
      vi.advanceTimersByTime(1);
      flushSync();
      expect(underlying.value).toBe("ab");
      expect(store.value).toBe("ab");
    });
    cleanup();
    vi.useRealTimers();
  });

  it("immediateSetter bypasses the debounce", () => {
    vi.useFakeTimers();
    const cleanup = $effect.root(() => {
      const underlying = new InMemoryRuneStore<string>("");
      const store = new DebouncedRuneStore<string>(underlying, 500);
      flushSync();

      store.immediateSetter("now");
      flushSync();
      expect(underlying.value).toBe("now");
      expect(store.value).toBe("now");
    });
    cleanup();
    vi.useRealTimers();
  });
});
