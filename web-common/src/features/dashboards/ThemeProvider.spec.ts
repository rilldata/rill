import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
    const fetchMock = vi.fn(
      () => new Promise((resolve) => (finishDownload = resolve)),
    );
    vi.stubGlobal("fetch", fetchMock);
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

    const project = { host: "http://localhost:9009", instanceId: "default" };
    const theme = new Theme(
      {
        font: {
          family: "Provider Sans",
          files: [{ path: "public/ProviderSans.woff2", weightMin: 400 }],
        },
      },
      project,
    );
    render(ThemeChartFontHarness, {
      props: { theme },
      context: new Map([[RUNTIME_CONTEXT_KEY, new RuntimeClient(project)]]),
    });

    const chartFont = screen.getByTestId("chart-font");
    expect(chartFont.textContent).toBe("default");
    // HTML text does not wait: the browser swaps the font in by itself.
    expect(chartFont.closest(".dashboard-theme-boundary")?.innerHTML).toContain(
      "font-family: var(--theme-font-family);",
    );

    // The file is requested once the client has checked that its token is fresh.
    await waitFor(() => expect(fetchMock).toHaveBeenCalledOnce());
    finishDownload({
      ok: true,
      arrayBuffer: () => Promise.resolve(new ArrayBuffer(8)),
    });

    await waitFor(() =>
      expect(chartFont.textContent).toBe(theme.font?.cssFamily),
    );
    expect(chartFont.textContent).toMatch(
      /^"rill-theme-font-\d+", "Provider Sans"/,
    );
  });
});
