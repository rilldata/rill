/** @type {import('tailwindcss').Config} */

// const primaryColors = {
// 		50: "#ecfff1",
// 		100: "#ddffe3",
// 		200: "#c2ffc5",
// 		300: "#9cffa4",
// 		400: "#75ff75",
// 		500: "#55ff69",
// 		600: "#35f535",
// 		700: "#35c73f",
// 		800: "#25ae2e",
// 		900: "#268933",
// 		950: "#195016",
// 	}

// const secondaryColors = {
// 	50: "#effaff",
// 	100: "#def3ff",
// 	200: "#b6e9ff",
// 	300: "#75daff",
// 	400: "#2cc9ff",
// 	500: "#00b8ff",
// 	600: "#008fd4",
// 	700: "#0071ab",
// 	800: "#00608d",
// 	900: "#065074",
// 	950: "#04324d",
// };

  /*
    colors for greyed-out elements. For now, using tailwind's
    standard "Gray", but using semantic color vars will
    allow us to change this to a custom palette
    or use e.g. tailwind's "slate", "zinc", etc if we want.

    Copied from https://github.com/shadcn-ui/ui/issues/669#issue-1771280130

    Visit that link if we want to copy/paste if
    we switch to "slate", "zinc", etc
    */

const mutedColors = {
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


// backup pallette of primary colors (green spectrum),
// useful for testing application of colors
const primaryColors = Object.fromEntries([50,100,200,300,400,500,600,700,800,900,950].map((n) => [n, `lch(${100*(1000-n)/1000}% 64 139)`]));

// backup pallette of secondary colors (red spectrum),
// useful for testing application of colors
const secondaryColors = Object.fromEntries([50,100,200,300,400,500,600,700,800,900,950].map((n) => [n, `lch(${100*(1000-n)/1000}% 78 13)`]));

	/**
	 * This function takes a color name and a color map, of the form
	 * {[colorKey]: colorValue}
	 * and returns an expanded map that has k/v pairs of the form e.g.
	 * 
	 * {`${colorName}-${key}`: colorValue}
	 * 
	 * I don't know exactly why this is needed, but when using tailwind's
	 * object syntax for colors:
	 * 
	 * ```
	 * primary: {
	 *     50: "#ecfff1",
	 *     ...
	 * }
	 * ```
	 * 
	 * of necessary classes were not being generated (`text-primary-800` etc)
	 * 
	 * @param {*} colorName 
	 * @param {*} coloMap 
	 */
	function expandColors(colorName, colorMap) {
		return Object.fromEntries(Object.entries(colorMap).map(([key, value]) => {
			return [`${colorName}-${key}`, value];
		}));
		
	}



module.exports = {
   // need to add this for storybook
  // https://www.kantega.no/blogg/setting-up-storybook-7-with-vite-and-tailwind-css
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx,svelte}",
  ],
  /** Once we have applied dark styling to all UI elements, remove this line */
  darkMode: "class",
  theme: {
    extend: {
      colors: {
				...expandColors('primary', primaryColors),
				...expandColors('secondary', secondaryColors),
				...expandColors('muted', mutedColors),
        border: "hsl(var(--border) / <alpha-value>)",
				input: "hsl(var(--input) / <alpha-value>)",
				ring: "hsl(var(--ring) / <alpha-value>)",
				background: "hsl(var(--background) / <alpha-value>)",
				foreground: "hsl(var(--foreground) / <alpha-value>)",
				primary: {
					DEFAULT: "hsl(var(--primary) / <alpha-value>)",
					foreground: "hsl(var(--primary-foreground) / <alpha-value>)"
				},
				secondary: {
					DEFAULT: "hsl(var(--secondary) / <alpha-value>)",
					foreground: "hsl(var(--secondary-foreground) / <alpha-value>)"
				},
				destructive: {
					DEFAULT: "hsl(var(--destructive) / <alpha-value>)",
					foreground: "hsl(var(--destructive-foreground) / <alpha-value>)"
				},
				muted: {
					DEFAULT: "hsl(var(--muted) / <alpha-value>)",
					foreground: "hsl(var(--muted-foreground) / <alpha-value>)"
				},
				accent: {
					DEFAULT: "hsl(var(--accent) / <alpha-value>)",
					foreground: "hsl(var(--accent-foreground) / <alpha-value>)"
				},
				popover: {
					DEFAULT: "hsl(var(--popover) / <alpha-value>)",
					foreground: "hsl(var(--popover-foreground) / <alpha-value>)"
				},
				card: {
					DEFAULT: "hsl(var(--card) / <alpha-value>)",
					foreground: "hsl(var(--card-foreground) / <alpha-value>)"
				}
      },
      borderRadius: {
				lg: "var(--radius)",
				md: "calc(var(--radius) - 2px)",
				sm: "calc(var(--radius) - 4px)"
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
		function({ addBase, theme }) {
      function extractColorVars(colorObj, colorGroup = '') {
        return Object.keys(colorObj).reduce((vars, colorKey) => {
          const value = colorObj[colorKey];
          const newVars =
            typeof value === 'string'
              ? { [`--color${colorGroup}-${colorKey}`]: value }
              : extractColorVars(value, `-${colorKey}`);

          return { ...vars, ...newVars };
        }, {});
      }

      addBase({
        ':root': extractColorVars(theme('colors')),
      });
    },
	],
  safelist: [
		// FIXME: i think this can come out now, because we've added
		// the "primary-800", so `text-`, `bg-` ect should be generated
    "text-primary-800", // needed for blue text in filter pills
    "ui-copy-code", // needed for code in measure expressions
  ],
};
