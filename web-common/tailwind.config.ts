import type { Config } from "tailwindcss";
import {
  type LightnessMap,
  type ThemeColorKind,
  defaultSecondaryColors,
  mutedColors,
} from "./src/features/themes/color-config";

/**
 * Takes a LightnessMap map object and a ThemeColorKind
 * ("primary" | "secondary" | "muted"),
 * and returns an object of the form e.g:
 * {
 *   "primary-50": "var(--color-primary-50)",
 *   "primary-100": "var(--color-primary-100)",
 * ...
 * }
 */

const tailwindScale = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950];
function addThemeColorsAsVarRefs(themeColorKind: string) {
  return Object.fromEntries(
    tailwindScale.map((colorNum) => {
      return [
        `${colorNum}`,
        `hsl(var(--hsl-${themeColorKind}-${colorNum}) / <alpha-value>)`,
      ];
    }),
  );
}

/**
 * Takes a LightnessMap map object and a ThemeColorKind
 * ("primary" | "secondary" | "muted"),
 * and returns an object of the form e.g:
 * {
 *   "--color-primary-50": "#ecf0ff",
 *   "--color-primary-100": "#dde4ff",
 * ...
 * }
 */
function initializeDefaultColorsVars(
  colorMap: LightnessMap,
  colorName: ThemeColorKind,
) {
  const colorVars: [string, string][] = Object.entries(colorMap).map(
    ([colorNum, colorCssString]) => [
      `--color-${colorName}-${colorNum}`,
      `hsl(${colorCssString})`,
    ],
  );
  const rawHSL: [string, string][] = Object.entries(colorMap).map(
    ([colorNum, colorCssString]) => [
      `--hsl-${colorName}-${colorNum}`,
      colorCssString,
    ],
  );
  return Object.fromEntries([...colorVars, ...rawHSL]);
}

export default {
  // need to add this for storybook
  // https://www.kantega.no/blogg/setting-up-storybook-7-with-vite-and-tailwind-css
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx,svelte}"],
  /** Once we have applied dark styling to all UI elements, remove this line */
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border) / <alpha-value>)",
        input: "hsl(var(--input) / <alpha-value>)",
        ring: "hsl(var(--ring) / <alpha-value>)",
        background: "hsl(var(--background) / <alpha-value>)",
        foreground: "hsl(var(--foreground) / <alpha-value>)",
        primary: {
          DEFAULT: "hsl(var(--primary) / <alpha-value>)",
          foreground: "hsl(var(--primary-foreground) / <alpha-value>)",
          ...addThemeColorsAsVarRefs("primary"),
        },
        white: {
          DEFAULT: "hsl(var(--background) / <alpha-value>)",
        },
        gray: {
          DEFAULT: "hsl(var(--primary) / <alpha-value>)",
          foreground: "hsl(var(--primary-foreground) / <alpha-value>)",
          ...addThemeColorsAsVarRefs("gray"),
        },

        slate: {
          DEFAULT: "hsl(var(--primary) / <alpha-value>)",
          foreground: "hsl(var(--primary-foreground) / <alpha-value>)",
          ...addThemeColorsAsVarRefs("slate"),
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary) / <alpha-value>)",
          foreground: "hsl(var(--secondary-foreground) / <alpha-value>)",
          ...addThemeColorsAsVarRefs("secondary"),
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive) / <alpha-value>)",
          foreground: "hsl(var(--destructive-foreground) / <alpha-value>)",
        },
        muted: {
          DEFAULT: "hsl(var(--muted) / <alpha-value>)",
          foreground: "hsl(var(--muted-foreground) / <alpha-value>)",
          ...addThemeColorsAsVarRefs("muted"),
        },
        accent: {
          DEFAULT: "hsl(var(--accent) / <alpha-value>)",
          foreground: "hsl(var(--accent-foreground) / <alpha-value>)",
        },
        popover: {
          DEFAULT: "hsl(var(--popover) / <alpha-value>)",
          foreground: "hsl(var(--popover-foreground) / <alpha-value>)",
        },
        card: {
          DEFAULT: "hsl(var(--card) / <alpha-value>)",
          foreground: "hsl(var(--card-foreground) / <alpha-value>)",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
    },
  },
  plugins: [
    function ({ addBase }) {
      const colorVars = {
        // ...initializeDefaultColorsVars(defaultPrimaryColors, "primary"),
        ...initializeDefaultColorsVars(defaultSecondaryColors, "secondary"),
        ...initializeDefaultColorsVars(mutedColors, "muted"),
      };
      addBase({
        ":root": colorVars,
      });
    },
  ],
  safelist: [
    "ui-copy-code", // needed for code in measure expressions
  ],
} satisfies Config;
