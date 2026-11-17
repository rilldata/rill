import { RUNTIME_CONTEXT_KEY } from "@rilldata/web-common/runtime-client/v2/context";
import { render, screen, waitFor } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Theme } from "../themes/theme";
import ThemeChartFontHarness from "./__fixtures__/ThemeChartFontHarness.svelte";

describe("ThemeProvider", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("hands charts the theme font only once its files have loaded", async () => {
    // Vega measures labels as it lays a chart out, so a chart that saw the family before the
    // browser could draw with it would keep the fallback font's metrics.
    let finishDownload: (response: unknown) => void = () => {};
    vi.stubGlobal(
      "fetch",
      vi.fn(() => new Promise((resolve) => (finishDownload = resolve))),
    );
    vi.stubGlobal(
      "FontFace",
      class {
        load = () => Promise.resolve(this);
      },
    );
    Object.defineProperty(document, "fonts", {
      configurable: true,
      value: { add: vi.fn(), delete: vi.fn() },
    });

    const theme = new Theme({
      font: {
        family: "Provider Sans",
        files: [{ path: "public/ProviderSans.woff2", weightMin: 400 }],
      },
    });
    render(ThemeChartFontHarness, {
      props: { theme },
      context: new Map([
        [
          RUNTIME_CONTEXT_KEY,
          {
            host: "http://localhost:9009",
            instanceId: "default",
            getJwt: () => undefined,
          },
        ],
      ]),
    });

    const chartFont = screen.getByTestId("chart-font");
    expect(chartFont.textContent).toBe("default");
    // HTML text does not wait: the browser swaps the font in by itself.
    expect(chartFont.closest(".dashboard-theme-boundary")?.innerHTML).toContain(
      "font-family: var(--theme-font-family);",
    );

    finishDownload({
      ok: true,
      arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)),
    });

    await waitFor(() =>
      expect(chartFont.textContent).toBe(
        `"Provider Sans", "Inter", sans-serif`,
      ),
    );
  });
});
