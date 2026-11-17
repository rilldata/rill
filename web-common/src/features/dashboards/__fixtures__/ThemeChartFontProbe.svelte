<script lang="ts">
  import { getThemeChartFont } from "../../themes/theme-font";

  const chartFont = getThemeChartFont();
</script>

<span data-testid="chart-font">{$chartFont ?? "default"}</span>
