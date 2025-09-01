import "@testing-library/jest-dom";
import { vi } from "vitest";
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
