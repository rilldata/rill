/**
 * This type represents a color palette, where the keys are
 * the color lightness numbers (50, 100, 200, etc) and the
 * values are css color strings.
 */
export type LightnessMap = { [key: number]: string };

/**
 * The three categories of colors that could be rethemed.
 * (though we only support primary and secondary for now)
 */
export type ThemeColorKind = "primary" | "secondary" | "muted";

// NOTE: run-time defaultPrimaryColors here MUST match
// compile-time defaultPrimaryColors in
// web-common/tailwind.config.cjs.
//
// Runtime copy is needed because components can load
// before css is loaded, so we need a copy in the bundle.
// Compile-time copy is needed so that tailwind can generate
// color classes for the default colors (bg-primary-500, etc).
export const defaultPrimaryColors: LightnessMap = {
  50: "#ecf0ff",
  100: "#dde4ff",
  200: "#c2cdff",
  300: "#9cabff",
  400: "#757eff",
  500: "#5655ff",
  600: "#4735f5",
  700: "#3542c7",
  800: "#3125ae",
  900: "#2c2689",
  950: "#1c1650",
};

// export const defaultPrimaryColors = Object.fromEntries(
//   [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950].map((n) => [
//     n,
//     `lch(${(100 * (1000 - n)) / 1000}% 64 139)`,
//   ]),
// );

export const defaultSecondaryColors = {
  50: "#effaff",
  100: "#def3ff",
  200: "#b6e9ff",
  300: "#75daff",
  400: "#2cc9ff",
  500: "#00b8ff",
  600: "#008fd4",
  700: "#0071ab",
  800: "#00608d",
  900: "#065074",
  950: "#04324d",
};

// backup pallette of secondary colors (red spectrum),
// useful for testing application of colors
// export const defaultSecondaryColors = Object.fromEntries(
//   [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950].map((n) => [
//     n,
//     `lch(${(100 * (1000 - n)) / 1000}% 78 13)`,
//   ]),
// );

/*
	colors for greyed-out elements. For now, using tailwind's
	standard "Gray", but using semantic color vars will
	allow us to change this to a custom palette
	or use e.g. tailwind's "slate", "zinc", etc if we want.

	Copied from https://github.com/shadcn-ui/ui/issues/669#issue-1771280130

	Visit that link if we want to copy/paste if
	we switch to "slate", "zinc", etc
	*/
export const mutedColors = {
  50: "hsl(210, 20%, 98%)",
  100: "hsl(220, 14.3%, 95.9%)",
  200: "hsl(220, 13%, 91%)",
  300: "hsl(216, 12.2%, 83.9%)",
  400: "hsl(217.9, 10.6%, 64.9%)",
  500: "hsl(220, 8.9%, 46.1%)",
  600: "hsl(215, 13.8%, 34.1%)",
  700: "hsl(216.9, 19.1%, 26.7%)",
  800: "hsl(215, 27.9%, 16.9%)",
  900: "hsl(220.9, 39.3%, 11%)",
  950: "hsl(224, 71.4%, 4.1%);",
};
