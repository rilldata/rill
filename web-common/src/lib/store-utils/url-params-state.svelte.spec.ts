import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { goto } from "$app/navigation";
import { page } from "$app/state";
import { ArrayRuneStore } from "./types.svelte.ts";
import { UrlParamsState } from "./url-params-state.svelte.ts";

vi.mock("$app/navigation", () => ({ goto: vi.fn() }));
vi.mock("$app/state", () => ({ page: { url: new URL("http://localhost/") } }));

function lastGotoUrl() {
  const calls = vi.mocked(goto).mock.calls;
  return calls[calls.length - 1][0] as URL;
}

describe("UrlParamsState", () => {
  beforeEach(() => {
    page.url = new URL("http://localhost/");
  });

  afterEach(async () => {
    // Flush any params queued by a test so they don't leak into the next one.
    await tick();
  });

  it("initializes its value from the URL param", () => {
    page.url = new URL("http://localhost/?q=hello&status=a,b");
    const cleanup = $effect.root(() => {
      const q = UrlParamsState.createStringParam("q");
      const status = UrlParamsState.createStringArrayParam("status");
      expect(q.value).toBe("hello");
      expect(status.value).toEqual(["a", "b"]);
    });
    cleanup();
  });

  it("falls back to the default when the param is absent", () => {
    const cleanup = $effect.root(() => {
      const q = UrlParamsState.createStringParam("q", "fallback");
      const status = UrlParamsState.createStringArrayParam("status");
      expect(q.value).toBe("fallback");
      expect(status.value).toEqual([]);
    });
    cleanup();
  });

  it("updates the value optimistically before the URL write is flushed", async () => {
    let q!: UrlParamsState<string, null>;
    const cleanup = $effect.root(() => {
      q = UrlParamsState.createStringParam("q");
    });

    q.setter("hello");
    // Value is available synchronously; the navigation is deferred.
    expect(q.value).toBe("hello");
    expect(q.getter()).toBe("hello");
    expect(goto).not.toHaveBeenCalled();

    await tick();
    expect(goto).toHaveBeenCalledTimes(1);
    expect(lastGotoUrl().searchParams.get("q")).toBe("hello");
    expect(vi.mocked(goto).mock.calls[0][1]).toEqual({
      noScroll: true,
      keepFocus: true,
    });
    cleanup();
  });

  it("batches writes from multiple stores into a single navigation", async () => {
    let q!: UrlParamsState<string, null>;
    let status!: ArrayRuneStore<string>;
    const cleanup = $effect.root(() => {
      q = UrlParamsState.createStringParam("q");
      status = UrlParamsState.createStringArrayParam("status");
    });

    q.setter("hello");
    status.setter(["a", "b"]);
    await tick();

    expect(goto).toHaveBeenCalledTimes(1);
    const url = lastGotoUrl();
    expect(url.searchParams.get("q")).toBe("hello");
    expect(url.searchParams.get("status")).toBe("a,b");
    cleanup();
  });

  it("removes the param when set to the empty/default value", async () => {
    page.url = new URL("http://localhost/?q=hello");
    let q!: UrlParamsState<string, null>;
    const cleanup = $effect.root(() => {
      q = UrlParamsState.createStringParam("q");
    });

    q.setter("");
    expect(q.value).toBe("");
    await tick();
    expect(lastGotoUrl().searchParams.has("q")).toBe(false);
    cleanup();
  });

  it("accumulates multiple toggles on the same store within one tick", async () => {
    // Regression: the optimistic value update lets the second toggle read the
    // first toggle's result instead of the stale URL value.
    const cleanup = $effect.root(() => {
      const status = UrlParamsState.createStringArrayParam("status");
      status.toggle("a");
      status.toggle("b");
      // Read synchronously, before any effect flush: the optimistic value lets
      // the second toggle build on the first instead of the stale URL value.
      expect(status.value).toEqual(["a", "b"]);
    });

    await tick();
    expect(lastGotoUrl().searchParams.get("status")).toBe("a,b");
    cleanup();
  });
});
