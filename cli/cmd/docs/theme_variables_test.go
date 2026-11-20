package docs

import (
	"testing"

	"github.com/rilldata/rill/runtime/parser"
	"github.com/stretchr/testify/require"
)

func TestThemeStylesDefaultFor(t *testing.T) {
	styles := &themeStyles{light: make(map[string]string), dark: make(map[string]string)}
	styles.add(`
@layer base {
  :root {
    /* Palette */
    --white: oklch(100% 0 0);
    --color-gray-light-50: oklch(0.985 0.002 247.839);
    --color-gray-dark-50: #101828;
    --color-gray-50: var(--color-gray-light-50);
    --color-neutral-light-100: oklch(0.97 0 none);
    --color-primary-light-50: oklch(0.96 0.019 273.047);
    --color-primary-light-100: oklch(0.93 0.034 272.788);
    --color-primary-light-300: oklch(0.785 0.115 274.713);
    --color-primary-50: var(--color-primary-light-50);
    --color-theme-300: var(
      --color-theme-light-300,
      var(--color-primary-light-300)
    );

    /* Semantic variables */
    --surface-subtle: var(--color-gray-50);
    --surface-muted: var(--color-neutral-light-100);
    --surface-hover: var(--color-primary-50);
    --popover-accent: var(--surface-hover);
    --fg-accent: var(--color-primary-light-300);
    --ring-focus: var(--color-theme-300);
    --input: color-mix(in oklab, var(--white) calc(0.05 * 100%), transparent);
    --border: hsl(0deg 0% 90%);
    --radius: 0.5rem;
  }

  :root.dark {
    --color-gray-50: var(--color-gray-dark-50);
    --fg-accent: var(--color-primary-light-100);
    --border: color-mix(
      in oklab,
      var(--white) calc(0.08 * 100%),
      transparent
    );
  }
}
`)

	tests := []struct {
		variable parser.ThemeVariable
		light    string
		dark     string
	}{
		// A palette color, which dark mode swaps through the palette alias.
		{parser.ThemeVariable{Name: "surface-subtle"}, "`#f9fafb`", "`#101828`"},
		// A gray without a hue.
		{parser.ThemeVariable{Name: "surface-muted"}, "`#f5f5f5`", "`#f5f5f5`"},
		// Shades of the brand palettes are named instead of resolved, since they follow the theme's own `primary`.
		{parser.ThemeVariable{Name: "surface-hover"}, "Shade 50 of `primary`", "Shade 50 of `primary`"},
		{parser.ThemeVariable{Name: "popover-accent"}, "Shade 50 of `primary`", "Shade 50 of `primary`"},
		{parser.ThemeVariable{Name: "fg-accent"}, "Shade 300 of `primary`", "Shade 100 of `primary`"},
		// The theme palette falls back to the primary palette.
		{parser.ThemeVariable{Name: "ring-focus"}, "Shade 300 of `primary`", "Shade 300 of `primary`"},
		// An opacity applied with color-mix(), on one line and wrapped over several.
		{parser.ThemeVariable{Name: "input"}, "`#ffffff` at 5% opacity", "`#ffffff` at 5% opacity"},
		{parser.ThemeVariable{Name: "border"}, "`#e6e6e6`", "`#ffffff` at 8% opacity"},
		{parser.ThemeVariable{Name: "radius", Length: true}, "`0.5rem`", "`0.5rem`"},
	}
	for _, tt := range tests {
		t.Run(tt.variable.Name, func(t *testing.T) {
			light, err := styles.defaultFor(tt.variable, false)
			require.NoError(t, err)
			require.Equal(t, tt.light, light)

			dark, err := styles.defaultFor(tt.variable, true)
			require.NoError(t, err)
			require.Equal(t, tt.dark, dark)
		})
	}

	_, err := styles.defaultFor(parser.ThemeVariable{Name: "tooltip"}, false)
	require.ErrorContains(t, err, "not declared")
}
