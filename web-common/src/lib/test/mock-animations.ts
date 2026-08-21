import { afterAll, beforeAll, vi } from "vitest";

export function mockAnimationsForComponentTesting() {
  // There is some weirdness with jsdom and svelte-transitions.
  // Mocking requestAnimationFrame like this produces better results.
  // Ref: https://github.com/testing-library/svelte-testing-library/issues/206#issuecomment-1470158576
  beforeAll(() => {
    vi.stubGlobal("requestAnimationFrame", (fn) => {
      return window.setTimeout(() => fn(Date.now()), 1);
    });

    // Svelte 5 transitions drive themselves with the Web Animations API, which jsdom does not
    // implement. Without this, a component that transitions throws "element.animate is not a
    // function" from a task that no test can catch.
    if (!Element.prototype.animate) {
      Element.prototype.animate = () => {
        const animation = {
          currentTime: 0,
          // Reports the animation as done so that the transition does not start a tick loop.
          playState: "finished",
          effect: null,
          onfinish: null as (() => void) | null,
          cancel: () => {},
        };
        // The caller assigns `onfinish` right after this returns, so it needs a tick to get there.
        window.setTimeout(() => animation.onfinish?.());
        return animation as unknown as Animation;
      };
    }
  });

  afterAll(() => {
    vi.unstubAllGlobals();
  });
}
