<script lang="ts">
  import type { Theme } from "../../themes/theme";
  import ThemeProvider from "../ThemeProvider.svelte";
  import ThemeChartFontProbe from "./ThemeChartFontProbe.svelte";

  export let theme: Theme | undefined;
</script>

<ThemeProvider {theme}>
  <ThemeChartFontProbe />
</ThemeProvider>
