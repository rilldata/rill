/** This test suite mocks the context functionality in Svelte.
 * A context really is just a map that spans across components on instantiation.
 * Thus it's easy to just mock the context functionality with a simple Map.
 * See https://github.com/sveltejs/svelte/blob/0f94c890f5fde899c40f2be05bce8e87579f26f3/src/runtime/internal/lifecycle.ts#L69
 * for details (you can fish around the Svelte repository to go deeper on contexts if needed).
 * This also helps us isolate the "cascading" part of the cascadingContextStore.
 *
 * Once we have stronger component level tests, we should rewrite this test suite to render svelte components.
 */
import { get } from "svelte/store";
import { cascadingContextStore } from "./cascading-context-store";
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import { getContext, hasContext, setContext } from "svelte";

let CONTEXT = new Map();
jest.mock("svelte", () => ({
  getContext(key: string) {
    return CONTEXT.get(key);
  },
  setContext(key: string, value: unknown) {
    CONTEXT.set(key, value);
  },
  hasContext(key: string) {
    return CONTEXT.has(key);
  },
}));

describe("cascadingContextStore", () => {
  beforeEach(() => {
    CONTEXT = new Map();
  });
  it("instantiates an empty cascading context store, with no parent inheritance", () => {
    const store = cascadingContextStore("test1", { a: 10, b: 20 });
    expect(get(store)).toEqual({ a: 10, b: 20 });
    expect(store.hasParentCascade).toBeFalsy();
  });
  it("overrides an old cascading context store", () => {
    const store = cascadingContextStore("test1", { a: 10, b: 20 });
    const store2 = cascadingContextStore("test1", { a: 30, c: 1000 });
    expect(get(store)).toEqual({ a: 10, b: 20 });
    expect(get(store2)).toEqual({ a: 30, b: 20, c: 1000 });
    expect(store.hasParentCascade).toBeFalsy();
    expect(store2.hasParentCascade).toBeTruthy();
  });
  it("reconciles props from parent and those passed in", () => {
    const store = cascadingContextStore("test1", { a: 10, b: 20 });
    const store2 = cascadingContextStore("test1", { a: 30, c: 1000 });
    // update the parent.
    store.reconcileProps({ b: 100, a: 10 });
    expect(get(store)).toEqual({ a: 10, b: 100 });
    expect(get(store2)).toEqual({ a: 30, b: 100, c: 1000 });
    // update the child.
    store2.reconcileProps({ a: 50, c: 2000 });
    expect(get(store)).toEqual({ a: 10, b: 100 });
    expect(get(store2)).toEqual({ a: 50, b: 100, c: 2000 });
    // update the parent again, and see how the updates persist.
    store.reconcileProps({ a: 1, b: 1 });
    expect(get(store)).toEqual({ a: 1, b: 1 });
    expect(get(store2)).toEqual({ a: 50, b: 1, c: 2000 });
  });
  it("generates derived values", () => {
    const store = cascadingContextStore(
      "test1",
      { a: 10, b: 20 },
      { t1: (s) => s.a + s.b }
    );
    const store2 = cascadingContextStore(
      "test1",
      { a: 10, c: 30 },
      { t2: (s) => s.a + s.c }
    );
    expect(get(store)).toEqual({ a: 10, b: 20, t1: 30 });
    // it maintains the derivation just from the parent, unless overridden explicitly.
    expect(get(store2)).toEqual({ a: 10, b: 20, c: 30, t1: 30, t2: 40 });
    const store3 = cascadingContextStore(
      "test1",
      { c: 40, d: 40 },
      {
        t1: (s) => s.a + s.b,
        t3: (s) => s.c + s.d,
      }
    );
    expect(get(store3)).toEqual({
      a: 10,
      b: 20,
      c: 40,
      d: 40,
      t1: 30,
      t2: 40,
      t3: 80,
    });
  });
});
