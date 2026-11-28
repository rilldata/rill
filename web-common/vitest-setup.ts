import "@testing-library/jest-dom";
import { afterAll, vi } from "vitest";
import { Settings } from "luxon";

// required for svelte5 + jsdom as jsdom does not support matchMedia
Object.defineProperty(window, "matchMedia", {
  writable: true,
  enumerable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

Object.defineProperty(window, "scrollTo", {
  writable: true,
  enumerable: true,
  value: vi.fn(),
});

Settings.defaultWeekSettings = {
  minimalDays: 4,
  firstDay: 1,
  weekend: [6, 7],
};

// bits-ui locks <body> while a dialog or menu is open,
// and restores its style on a 24ms timer after the last one closes.
// Let that timer run before the jsdom environment is torn down;
// otherwise it fires without a `document` and Vitest fails the whole run.
afterAll(async () => {
  if (document.body.style.pointerEvents !== "none") return;
  await new Promise((resolve) => setTimeout(resolve, 50));
});
