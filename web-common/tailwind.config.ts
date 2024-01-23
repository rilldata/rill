import type { Config } from "tailwindcss";
import {
  LightnessMap,
  ThemeColorKind,
  defaultPrimaryColors,
  defaultSecondaryColors,
  mutedColors,
} from "./src/features/themes/color-config";

/**
 *
 * @param colorMap
 * @param colorName
 * @returns
 */
function addThemeColorsAsVarRefs(
  colorMap: LightnessMap,
  themeColorKind: ThemeColorKind,
) {
  return Object.fromEntries(
    Object.keys(colorMap).map((colorNum) => {
      return [
        `${themeColorKind}-${colorNum}`,
        `var(--color-${themeColorKind}-${colorNum})`,
      ];
    }),
  );
}

function initializeDefaultColorsVars(
  colorMap: LightnessMap,
  colorName: ThemeColorKind,
) {
  const colorVars: [string, string][] = Object.entries(colorMap).map(
    ([colorNum, colorCssString]) => [
      `--color-${colorName}-${colorNum}`,
      colorCssString,
    ],
  );
  return Object.fromEntries(colorVars);
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
        ...addThemeColorsAsVarRefs(defaultPrimaryColors, "primary"),
        ...addThemeColorsAsVarRefs(defaultSecondaryColors, "secondary"),
        ...addThemeColorsAsVarRefs(mutedColors, "muted"),

        border: "hsl(var(--border) / <alpha-value>)",
        input: "hsl(var(--input) / <alpha-value>)",
        ring: "hsl(var(--ring) / <alpha-value>)",
        background: "hsl(var(--background) / <alpha-value>)",
        foreground: "hsl(var(--foreground) / <alpha-value>)",
        primary: {
          DEFAULT: "hsl(var(--primary) / <alpha-value>)",
          foreground: "hsl(var(--primary-foreground) / <alpha-value>)",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary) / <alpha-value>)",
          foreground: "hsl(var(--secondary-foreground) / <alpha-value>)",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive) / <alpha-value>)",
          foreground: "hsl(var(--destructive-foreground) / <alpha-value>)",
        },
        muted: {
          DEFAULT: "hsl(var(--muted) / <alpha-value>)",
          foreground: "hsl(var(--muted-foreground) / <alpha-value>)",
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
    /**
     * Note: this plugin creates css variables for all colors
     * defined in the theme.colors object. These will be available
     * as e.g. `var(--color-COLOR_NAME-500)`.
     *
     * This allows us to define our colors in only this file,
     * without also needing to define them in the global CSS file.
     *
     * is taken from here:
     * https://gist.github.com/Merott/d2a19b32db07565e94f10d13d11a8574
     *
     */
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
