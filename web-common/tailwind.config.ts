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
        card: {
          DEFAULT: oklabString("card"),
          foreground: oklabString("card-foreground"),
        },
        popover: {
          DEFAULT: oklabString("popover"),
          foreground: oklabString("popover-foreground"),
          accent: oklabString("popover-accent"),
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
          primary: oklabString("accent-primary"),
          "primary-action": oklabString("accent-primary-action"),
          secondary: oklabString("accent-secondary"),
          "secondary-action": oklabString("accent-secondary-action"),
        },
        icon: {
          DEFAULT: oklabString("icon-default"),
          default: oklabString("icon-default"),
          muted: oklabString("icon-muted"),
          disabled: oklabString("icon-disabled"),
          accent: oklabString("icon-accent"),
        },
        destructive: {
          DEFAULT: oklabString("destructive"),
          foreground: oklabString("destructive-foreground"),
        },
        border: oklabString("border"),
        input: "var(--input)",
        ring: {
          DEFAULT: oklabString("ring"),
          focus: oklabString("ring-focus"),
          offset: oklabString("ring-offset"),
        },
        sidebar: {
          DEFAULT: oklabString("sidebar"),
          foreground: oklabString("sidebar-foreground"),
        },
        surface: {
          DEFAULT: oklabString("surface"),
          background: oklabString("surface-background"),
          container: oklabString("surface-container"),
          "container-hover": oklabString("surface-container-hover"),
          "container-active": oklabString("surface-container-active"),
          elevated: oklabString("surface-elevated"),
          overlay: oklabString("surface-overlay"),
          muted: oklabString("surface-muted"),
        },
        fg: {
          DEFAULT: oklabString("fg-primary"),
          primary: oklabString("fg-primary"),
          secondary: oklabString("fg-secondary"),
          tertiary: oklabString("fg-tertiary"),
          inverse: oklabString("fg-inverse"),
          muted: oklabString("fg-muted"),
          disabled: oklabString("fg-disabled"),
          accent: oklabString("fg-accent"),
        },
        theme: {
          DEFAULT: oklabString("color-theme-500"),
          foreground: oklabString("theme-foreground"),
          ...genColorObject("theme"),
        },
        Canvas: oklabString("canvas"),
        Explore: oklabString("explore"),
        Metrics: oklabString("metrics"),
        Model: oklabString("model"),
        API: oklabString("api"),
        Data: oklabString("data"),
        Theme: oklabString("theme"),
        Alert: oklabString("alert"),
        Report: oklabString("report"),
        Connector: oklabString("connector"),
        Component: oklabString("component"),
        dimension: {
          DEFAULT: oklabString("dimension"),
          foreground: oklabString("dimension-foreground"),
          border: oklabString("dimension-border"),
        },
        measure: {
          DEFAULT: oklabString("measure"),
          foreground: oklabString("measure-foreground"),
          border: oklabString("measure-border"),
        },
        tooltip: oklabString("tooltip"),

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
} satisfies Config;
