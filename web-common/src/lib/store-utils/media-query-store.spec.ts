import { get } from "svelte/store";
import { afterEach, describe, expect, it, vi } from "vitest";
import { mediaQueryStore } from "./media-query-store";

function mockMatchMedia(initiallyMatches: boolean) {
  const listeners = new Set<() => void>();
  const mediaQueryList = {
    matches: initiallyMatches,
    addEventListener: (_: string, listener: () => void) =>
      listeners.add(listener),
    removeEventListener: (_: string, listener: () => void) =>
      listeners.delete(listener),
  };
  vi.stubGlobal(
    "matchMedia",
    vi.fn(() => mediaQueryList),
  );
  return {
    listeners,
    change(matches: boolean) {
      mediaQueryList.matches = matches;
      listeners.forEach((listener) => listener());
    },
  };
}

describe("mediaQueryStore", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("starts with the current match", () => {
    mockMatchMedia(true);
    expect(get(mediaQueryStore("(max-width: 100px)"))).toBe(true);
  });

  it("follows changes and stops listening when unsubscribed", () => {
    const media = mockMatchMedia(false);
    const values: boolean[] = [];
    const unsubscribe = mediaQueryStore("(max-width: 100px)").subscribe((v) =>
      values.push(v),
    );
    media.change(true);
    expect(values).toEqual([false, true]);

    unsubscribe();
    expect(media.listeners.size).toBe(0);
  });
});
