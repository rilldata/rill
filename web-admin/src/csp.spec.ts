import { describe, expect, it } from "vitest";
import svelteConfig from "../svelte.config.js";

const directives = svelteConfig.kit.csp.directives;

// The canvas map component (web-common/src/features/canvas/components/map)
// uses Mapbox GL JS, which needs these directives to work under web-admin's
// CSP. See https://docs.mapbox.com/mapbox-gl-js/guides/browsers-and-testing/
describe("web-admin CSP", () => {
  it("allows Mapbox GL JS API requests", () => {
    expect(directives["connect-src"]).toEqual(
      expect.arrayContaining([
        "https://api.mapbox.com",
        "https://events.mapbox.com",
      ]),
    );
  });

  it("allows Mapbox GL JS blob: workers", () => {
    expect(directives["worker-src"]).toEqual(
      expect.arrayContaining(["self", "blob:"]),
    );
    expect(directives["child-src"]).toEqual(
      expect.arrayContaining(["self", "blob:"]),
    );
  });

  it("allows Mapbox GL JS tile and sprite images", () => {
    expect(directives["img-src"]).toEqual(
      expect.arrayContaining(["https:", "data:", "blob:"]),
    );
  });
});
