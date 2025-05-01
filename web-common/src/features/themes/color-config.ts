export const TailwindColorSpacing = [
  50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950,
] as const;

export const TailwindColors = [
  "red",
  "orange",
  "amber",
  "yellow",
  "lime",
  "green",
  "emerald",
  "teal",
  "cyan",
  "sky",
  "blue",
  "indigo",
  "violet",
  "purple",
  "fuchsia",
  "pink",
  "rose",
  "slate",
  "gray",
  "zinc",
  "neutral",
  "stone",
];

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

/**
 * Rill primary brand colors.
 */
export const defaultPrimaryColors: LightnessMap = {
  50: "227 100% 96%",
  100: "228 100% 93%",
  200: "229 100% 88%",
  300: "231 100% 81%",
  400: "236 100% 73%",
  500: "240 100% 67%",
  600: "246 91% 58%",
  700: "235 58% 49%",
  800: "245 65% 41%",
  900: "244 57% 34%",
  950: "246 57% 20%",
};

export const defaultSecondaryColors = {
  50: "199 100% 97%",
  100: "202 100% 94%",
  200: "198 100% 86%",
  300: "196 100% 73%",
  400: "195 100% 59%",
  500: "197 100% 50%",
  600: "200 100% 42%",
  700: "200 100% 34%",
  800: "199 100% 28%",
  900: "200 90% 24%",
  950: "202 90% 16%",
};

// backup pallette of secondary colors (red spectrum)",
// useful for testing application of colors
// export const defaultSecondaryColors = Object.fromEntries(
//   [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950].map((n) => [
//     n,
//     `lch(${(100 * (1000 - n)) / 1000}% 78 13)`,
//   ]),
// );

export function getRandomBgColor(name: string): string {
  const colorList = [
    "bg-blue-500",
    "bg-green-500",
    "bg-red-500",
    "bg-orange-500",
    "bg-yellow-500",
    "bg-amber-500",
    "bg-pink-500",
    "bg-lime-500",
    "bg-emerald-500",
    "bg-teal-500",
    "bg-cyan-500",
    "bg-sky-500",
    "bg-indigo-500",
    "bg-violet-500",
    "bg-purple-500",
    "bg-fuchsia-500",
    "bg-rose-500",
  ];

  if (!name) {
    return colorList[Math.floor(Math.random() * colorList.length)];
  }

  const hash = Array.from(name).reduce(
    (acc, char) => acc + char.charCodeAt(0),
    0,
  );
  const index = hash % colorList.length;

  return colorList[index];
}
