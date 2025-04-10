import { afterAll, beforeAll, vi } from "vitest";

export function mockAnimationsForComponentTesting() {
  // There is some weirdness with jsdom and svelte-transitions.
  // Mocking requestAnimationFrame like this produces better results.
  // Ref: https://github.com/testing-library/svelte-testing-library/issues/206#issuecomment-1470158576
  beforeAll(() => {
    vi.stubGlobal("requestAnimationFrame", (fn) => {
      return window.setTimeout(() => fn(Date.now()), 1);
    });
  });

  afterAll(() => {
    vi.unstubAllGlobals();
  });
}
