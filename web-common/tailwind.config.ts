import type { Config } from "tailwindcss";
import {
  type LightnessMap,
  TailwindColorSpacing,
  type ThemeColorKind,
  defaultPrimaryColors,
  TailwindColors,
  defaultSecondaryColors,
  mutedColors,
} from "./src/features/themes/color-config";

function addThemeColorsAsVarRefs(themeColorKind: string) {
  return Object.fromEntries(
    TailwindColorSpacing.map((colorNum) => {
      return [
        `${colorNum}`,
        `hsl(var(--hsl-${themeColorKind}-${colorNum}) / <alpha-value>)`,
      ];
    }),
  );
}

function generateTailwindVariables() {
  const colors: Record<string, Record<string, string>> = {};

  TailwindColors.forEach((color) => {
    colors[color] = genColorObject(color);
  });

  return colors;
}

function genColorObject(color: string) {
  return Object.fromEntries(
    TailwindColorSpacing.map((colorNum) => {
      return [
        `${colorNum}`,
        `color-mix(in oklab, var(--color-${color}-${colorNum}) calc(<alpha-value> * 100%), transparent)`,
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
        background: `var(--background)`,
        foreground: "hsl(var(--foreground) / <alpha-value>)",
        primary: {
          DEFAULT: "hsl(var(--primary) / <alpha-value>)",
          foreground: "hsl(var(--primary-foreground) / <alpha-value>)",
          ...genColorObject("primary"),
        },
        surface: "var(--surface)",

        ...generateTailwindVariables(),
        secondary: {
          DEFAULT: "hsl(var(--secondary) / <alpha-value>)",
          foreground: "hsl(var(--secondary-foreground) / <alpha-value>)",
          ...genColorObject("secondary"),
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive) / <alpha-value>)",
          foreground: "hsl(var(--destructive-foreground) / <alpha-value>)",
        },
        muted: {
          DEFAULT: `color-mix(in oklab, var(--muted) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--muted-foreground) calc(<alpha-value> * 100%), transparent)`,
        },
        accent: {
          DEFAULT: `color-mix(in oklab, var(--accent) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--accent-foreground) calc(<alpha-value> * 100%), transparent)`,
        },
        popover: "var(--popover)",
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
      zIndex: {
        popover: "80",
      },
    },
  },
  plugins: [
    function ({ addBase }) {
      const colorVars = {
        ...initializeDefaultColorsVars(defaultPrimaryColors, "primary"),
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
