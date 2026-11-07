import type { Config } from "tailwindcss";
import webCommonPreset from "../web-common/tailwind.config";

// The content list is explicit on purpose: the widget reuses only a handful of
// web-common components, and scanning all of web-common would ship unused utilities
// into the single-file bundle.
export default {
  presets: [webCommonPreset],
  content: [
    "./widget.html",
    "./src/**/*.{svelte,ts}",
    "../web-common/src/features/components/charts/Chart.svelte",
    "../web-common/src/components/vega/VegaLiteRenderer.svelte",
    "../web-common/src/features/components/ComponentError.svelte",
    "../web-common/src/features/entity-management/Spinner.svelte",
  ],
} satisfies Config;
