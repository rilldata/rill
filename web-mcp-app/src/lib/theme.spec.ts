import { updateAllPaletteColors } from "@rilldata/web-common/features/themes/palette-store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  applyHostTheme,
  resolveFontFamily,
  resolveIsDark,
  SYSTEM_FONT_STACK,
} from "./theme";

vi.mock(
  "@rilldata/web-common/features/themes/palette-store",
  async (importOriginal) => ({
    ...(await importOriginal<object>()),
    updateAllPaletteColors: vi.fn(),
  }),
);

describe("applyHostTheme", () => {
  beforeEach(() => {
    document.documentElement.className = "";
    document.documentElement.removeAttribute("data-theme");
    document.documentElement.removeAttribute("style");
  });

  it("toggles Rill's dark class and the ext-apps theme attribute", () => {
    const dark = applyHostTheme({ theme: "dark" });
    expect(dark.isDark).toBe(true);
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");

    const light = applyHostTheme({ theme: "light" });
    expect(light.isDark).toBe(false);
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  });

  it("applies host style variables and fonts", () => {
    applyHostTheme({
      theme: "light",
      styles: {
        variables: {
          "--font-sans": "Host Sans, sans-serif",
          "--color-text-primary": "#123456",
        } as never,
        css: { fonts: "@font-face { font-family: 'Host Sans'; src: url(x); }" },
      },
    });
    const root = document.documentElement;
    expect(root.style.getPropertyValue("--font-sans")).toBe(
      "Host Sans, sans-serif",
    );
    expect(root.style.getPropertyValue("--color-text-primary")).toBe("#123456");
    expect(document.getElementById("__mcp-host-fonts")?.textContent).toContain(
      "Host Sans",
    );
  });

  it("re-reads Rill's cached colors only when the mode or the host variables change", () => {
    const refreshes = vi.mocked(updateAllPaletteColors);
    const styles = { variables: { "--color-text-primary": "#111" } as never };
    applyHostTheme({ theme: "light", styles });
    const before = refreshes.mock.calls.length;

    // Resizes and other unrelated context changes do not re-read the colors.
    applyHostTheme({
      theme: "light",
      styles,
      containerDimensions: { width: 320, maxHeight: 400 },
    });
    applyHostTheme({
      theme: "light",
      styles,
      containerDimensions: { width: 640, maxHeight: 400 },
    });
    expect(refreshes.mock.calls.length).toBe(before);

    applyHostTheme({ theme: "dark", styles });
    expect(refreshes.mock.calls.length).toBe(before + 1);

    applyHostTheme({
      theme: "dark",
      styles: { variables: { "--color-text-primary": "#eee" } as never },
    });
    expect(refreshes.mock.calls.length).toBe(before + 2);
  });

  it("falls back to the system font stack", () => {
    expect(resolveFontFamily()).toBe(SYSTEM_FONT_STACK);
    expect(applyHostTheme(undefined).fontFamily).toBe(SYSTEM_FONT_STACK);
  });

  it("falls back to the media query when the host sends no theme", () => {
    expect(resolveIsDark(undefined)).toBe(false);
    expect(resolveIsDark({ theme: "dark" })).toBe(true);
  });
});
