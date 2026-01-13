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
      return [`${colorNum}`, oklabString(`color-${color}-${colorNum}`)];
    }),
  );
}

function oklabString(variableName: string) {
  return `color-mix(in oklab, var(--${variableName}) calc(<alpha-value> * 100%), transparent)`;
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
        DEFAULT: oklabString("border"),
      },
      colors: {
        background: oklabString("background"),
        foreground: oklabString("foreground"),
        neutral: {
          DEFAULT: oklabString("subtle"),
          foreground: oklabString("subtle-foreground"),
        },
        card: {
          DEFAULT: oklabString("card"),
          foreground: oklabString("card-foreground"),
        },
        popover: {
          DEFAULT: oklabString("popover"),
          foreground: oklabString("popover-foreground"),
          footer: oklabString("popover-footer"),
        },
        primary: {
          DEFAULT: oklabString("color-primary-500"),
          foreground: oklabString("color-gray-50"),
          ...genColorObject("primary"),
        },
        secondary: {
          DEFAULT: oklabString("color-secondary-500"),
          foreground: oklabString("color-gray-50"),
          ...genColorObject("secondary"),
        },
        muted: {
          DEFAULT: oklabString("muted"),
          foreground: oklabString("muted-foreground"),
        },
        accent: {
          DEFAULT: oklabString("accent"),
          foreground: oklabString("accent-foreground"),
        },
        destructive: {
          DEFAULT: oklabString("destructive"),
          foreground: oklabString("destructive-foreground"),
        },
        border: oklabString("border"),
        input: oklabString("input"),
        ring: oklabString("ring"),
        sidebar: {
          DEFAULT: oklabString("sidebar"),
          foreground: oklabString("sidebar-foreground"),
        },
        surface: {
          DEFAULT: oklabString("surface"),
          foreground: oklabString("surface-foreground"),
        },
        theme: {
          DEFAULT: oklabString("color-theme-500"),
          foreground: oklabString("theme-foreground"),
          ...genColorObject("theme"),
        },
        "theme-secondary": {
          DEFAULT: oklabString("color-theme-secondary-500"),
          foreground: oklabString("color-gray-50"),
          ...genColorObject("theme-secondary"),
        },
        ...generateTailwindVariables(),
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
