import type { Config } from "tailwindcss";
import {
  TailwindColorSpacing,
  TailwindColors,
} from "./src/features/themes/color-config";

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

export default {
  // need to add this for storybook
  // https://www.kantega.no/blogg/setting-up-storybook-7-with-vite-and-tailwind-css
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx,svelte}"],
  /** Once we have applied dark styling to all UI elements, remove this line */
  darkMode: "class",
  theme: {
    extend: {
      borderColor: {
        DEFAULT:
          "color-mix(in oklab, var(--border) calc(<alpha-value> * 100%), transparent)",
      },
      colors: {
        input:
          "color-mix(in oklab, var(--input) calc(<alpha-value> * 100%), transparent)",
        ring: "color-mix(in oklab, var(--ring) calc(<alpha-value> * 100%), transparent)",
        background: `var(--background)`,
        foreground:
          "color-mix(in oklab, var(--foreground) calc(<alpha-value> * 100%), transparent)",
        surface: "var(--surface)",
        popover: "var(--popover)",
        primary: {
          DEFAULT: `color-mix(in oklab, var(--color-primary-500) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--primary-foreground) calc(<alpha-value> * 100%), transparent)`,
          ...genColorObject("primary"),
        },
        theme: {
          DEFAULT: `color-mix(in oklab, var(--color-theme-500) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--theme-foreground) calc(<alpha-value> * 100%), transparent)`,
          ...genColorObject("theme"),
        },
        ...generateTailwindVariables(),
        secondary: {
          DEFAULT: `color-mix(in oklab, var(--color-secondary-500) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--secondary-foreground) calc(<alpha-value> * 100%), transparent)`,
          ...genColorObject("secondary"),
        },
        "theme-secondary": {
          DEFAULT: `color-mix(in oklab, var(--color-theme-secondary-500) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--secondary-foreground) calc(<alpha-value> * 100%), transparent)`,
          ...genColorObject("theme-secondary"),
        },
        destructive: {
          DEFAULT: `color-mix(in oklab, var(--destructive) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--destructive-foreground) calc(<alpha-value> * 100%), transparent)`,
        },
        muted: {
          DEFAULT: `color-mix(in oklab, var(--muted) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--muted-foreground) calc(<alpha-value> * 100%), transparent)`,
          ...genColorObject("muted"),
        },
        accent: {
          DEFAULT: `color-mix(in oklab, var(--accent) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--accent-foreground) calc(<alpha-value> * 100%), transparent)`,
        },

        card: {
          DEFAULT: `color-mix(in oklab, var(--card) calc(<alpha-value> * 100%), transparent)`,
          foreground: `color-mix(in oklab, var(--card-foreground) calc(<alpha-value> * 100%), transparent)`,
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

  safelist: [
    "ui-copy-code", // needed for code in measure expressions
  ],
} satisfies Config;
