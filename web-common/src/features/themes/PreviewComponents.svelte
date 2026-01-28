<script lang="ts">
  export let sequentialColors: string[];
  export let qualitativeColors: string[];
  export let divergingColors: string[];
  export let primaryColor: string;
  export let cardColor: string;
  export let fgPrimary: string;

  // Mock time series data for KPI sparkline
  const timeSeriesData = [
    { x: 0, y: 45 },
    { x: 1, y: 52 },
    { x: 2, y: 48 },
    { x: 3, y: 61 },
    { x: 4, y: 55 },
    { x: 5, y: 67 },
    { x: 6, y: 72 },
    { x: 7, y: 65 },
    { x: 8, y: 78 },
    { x: 9, y: 82 },
    { x: 10, y: 75 },
    { x: 11, y: 88 },
  ];

  // SVG viewBox dimensions for sparkline
  // Large width enables smooth scaling with preserveAspectRatio="none"
  const SPARKLINE_VIEWBOX_WIDTH = 10000;
  const SPARKLINE_VIEWBOX_HEIGHT = 100;
  const minY = Math.min(...timeSeriesData.map((d) => d.y));
  const maxY = Math.max(...timeSeriesData.map((d) => d.y));
  const yPadding = (maxY - minY) * 0.1;

  function scaleX(x: number): number {
    return (x / (timeSeriesData.length - 1)) * SPARKLINE_VIEWBOX_WIDTH;
  }

  function scaleY(y: number): number {
    return (
      SPARKLINE_VIEWBOX_HEIGHT -
      ((y - minY + yPadding) / (maxY - minY + yPadding * 2)) *
        SPARKLINE_VIEWBOX_HEIGHT
    );
  }

  $: linePath = timeSeriesData
    .map((point, i) => {
      const x = scaleX(point.x);
      const y = scaleY(point.y);
      return i === 0 ? `M ${x} ${y}` : `L ${x} ${y}`;
    })
    .join(" ");

  $: areaPath =
    linePath +
    ` L ${scaleX(timeSeriesData.length - 1)} ${SPARKLINE_VIEWBOX_HEIGHT} L ${scaleX(0)} ${SPARKLINE_VIEWBOX_HEIGHT} Z`;

  // Bar chart data
  const barData = [
    { label: "North America", value: 2400000 },
    { label: "Europe", value: 1800000 },
    { label: "Asia Pacific", value: 2100000 },
    { label: "Latin America", value: 950000 },
    { label: "Middle East", value: 720000 },
    { label: "Africa", value: 480000 },
  ];

  const maxBarValue = Math.max(...barData.map((d) => d.value));

  function formatValue(val: number): string {
    if (val >= 1000000) return `${(val / 1000000).toFixed(1)}M`;
    if (val >= 1000) return `${(val / 1000).toFixed(0)}K`;
    return val.toString();
  }

  // Bar chart dimensions
  const barChartWidth = 300;
  const barChartHeight = 140;
  const barChartPadding = { top: 10, right: 10, bottom: 24, left: 35 };
  const barAreaWidth =
    barChartWidth - barChartPadding.left - barChartPadding.right;
  const barAreaHeight =
    barChartHeight - barChartPadding.top - barChartPadding.bottom;

  // Sequential heatmap data (1-9 intensity scale)
  const seqRows = ["Mon", "Tue", "Wed", "Thu", "Fri"];
  const seqCols = ["6am", "9am", "12pm", "3pm", "6pm"];
  const seqData = [
    [2, 5, 7, 8, 6],
    [3, 6, 8, 9, 7],
    [2, 5, 7, 8, 6],
    [3, 6, 9, 8, 7],
    [4, 7, 8, 7, 5],
  ];

  // Diverging heatmap data (-5 to +5 change, mapped to 11 colors)
  // Index 0 = -5 (color 1), Index 5 = 0 (color 6), Index 10 = +5 (color 11)
  const divRows = ["Q1", "Q2", "Q3", "Q4"];
  const divCols = ["2021", "2022", "2023", "2024"];
  const divData = [
    [-4, -2, 1, 3],
    [-3, 0, 2, 4],
    [-2, 1, 3, 5],
    [-1, 2, 4, 5],
  ];

  // Map diverging value (-5 to +5) to color index (0-10)
  function divValueToColorIndex(value: number): number {
    return Math.max(0, Math.min(10, value + 5));
  }

  // Heatmap dimensions
  const heatmapWidth = 280;
  const heatmapHeight = 130;
  const heatmapPadding = { top: 20, right: 10, bottom: 10, left: 30 };

  $: seqCellWidth =
    (heatmapWidth - heatmapPadding.left - heatmapPadding.right) /
    seqCols.length;
  $: seqCellHeight =
    (heatmapHeight - heatmapPadding.top - heatmapPadding.bottom) /
    seqRows.length;

  $: divCellWidth =
    (heatmapWidth - heatmapPadding.left - heatmapPadding.right) /
    divCols.length;
  $: divCellHeight =
    (heatmapHeight - heatmapPadding.top - heatmapPadding.bottom) /
    divRows.length;

  const gradientId = `kpi-gradient-${Math.random().toString(36).substring(2, 11)}`;

  // Leaderboard data
  const leaderboardData = [
    { rank: 1, name: "Electronics", value: 482000 },
    { rank: 2, name: "Clothing", value: 356000 },
    { rank: 3, name: "Home & Garden", value: 289000 },
    { rank: 4, name: "Sports", value: 201000 },
    { rank: 5, name: "Books", value: 145000 },
  ];
  const maxLeaderboardValue = Math.max(...leaderboardData.map((d) => d.value));
</script>

<div class="grid grid-cols-1 gap-3">
  <!-- KPI Component -->
  <div class="preview-card kpi-card" style="background-color: {cardColor};">
    <div class="kpi-data-wrapper">
      <div class="kpi-header text-gray-600 dark:text-gray-400">Revenue</div>
      <div class="kpi-value" style="color: {fgPrimary};">$2.4M</div>
      <div class="kpi-comparison">
        <span class="kpi-prev text-gray-500">$2.1M</span>
        <span class="kpi-delta" style="color: {primaryColor};">+$285K</span>
        <span class="kpi-percent text-gray-500">+12.5%</span>
      </div>
      <div class="kpi-label text-gray-400">vs last month</div>
    </div>
    <div class="kpi-sparkline-wrapper">
      <svg
        class="kpi-sparkline-svg"
        viewBox="0 0 {SPARKLINE_VIEWBOX_WIDTH} {SPARKLINE_VIEWBOX_HEIGHT}"
        preserveAspectRatio="none"
      >
        <defs>
          <linearGradient id={gradientId} x1="0" x2="0" y1="0" y2="1">
            <stop offset="5%" stop-color={primaryColor} stop-opacity="0.3" />
            <stop offset="95%" stop-color={primaryColor} stop-opacity="0.05" />
          </linearGradient>
        </defs>
        <path d={areaPath} fill="url(#{gradientId})" />
        <path
          d={linePath}
          fill="none"
          stroke={primaryColor}
          stroke-width="1"
          vector-effect="non-scaling-stroke"
        />
      </svg>
      <div class="kpi-dates text-gray-500">
        <span>Nov 1</span>
        <span>Nov 30</span>
      </div>
    </div>
  </div>

  <!-- Leaderboard Component -->
  <div
    class="preview-card leaderboard-card"
    style="background-color: {cardColor};"
  >
    <div class="chart-title" style="color: {fgPrimary};">Top Categories</div>
    <table class="leaderboard-table">
      <tbody>
        {#each leaderboardData as item}
          {@const barLength = (item.value / maxLeaderboardValue) * 140}
          <tr class="leaderboard-row">
            <td
              class="leaderboard-name"
              style="background: linear-gradient(to right, var(--color-theme-100) {barLength}px, transparent {barLength}px);"
            >
              <span style="color: {fgPrimary};">{item.name}</span>
            </td>
            <td
              class="leaderboard-value"
              style="background: linear-gradient(to right, var(--color-theme-100) {Math.max(
                0,
                barLength - 90,
              )}px, transparent {Math.max(0, barLength - 90)}px);"
            >
              <span style="color: {fgPrimary};">{formatValue(item.value)}</span>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <!-- Bar Chart Component -->
  <div
    class="preview-card bar-chart-card"
    style="background-color: {cardColor};"
  >
    <div class="chart-title" style="color: {fgPrimary};">Sales by Region</div>
    <svg
      class="bar-chart-svg"
      viewBox="0 0 {barChartWidth} {barChartHeight}"
      preserveAspectRatio="xMidYMid meet"
    >
      {#each [0, 0.5, 1] as tick}
        <text
          x={barChartPadding.left - 4}
          y={barChartPadding.top + barAreaHeight * (1 - tick)}
          text-anchor="end"
          dominant-baseline="middle"
          class="axis-label"
        >
          {formatValue(maxBarValue * tick)}
        </text>
        <line
          x1={barChartPadding.left}
          y1={barChartPadding.top + barAreaHeight * (1 - tick)}
          x2={barChartPadding.left + barAreaWidth}
          y2={barChartPadding.top + barAreaHeight * (1 - tick)}
          class="grid-line"
        />
      {/each}

      {#each barData as item, i}
        {@const barWidth = (barAreaWidth / barData.length) * 0.7}
        {@const barGap = (barAreaWidth / barData.length) * 0.15}
        {@const barHeight = (item.value / maxBarValue) * barAreaHeight}
        {@const x =
          barChartPadding.left + i * (barAreaWidth / barData.length) + barGap}
        {@const y = barChartPadding.top + barAreaHeight - barHeight}
        <rect
          {x}
          {y}
          width={barWidth}
          height={barHeight}
          fill={qualitativeColors[i] || primaryColor}
          rx="2"
        />
        <text
          x={x + barWidth / 2}
          y={barChartHeight - 4}
          text-anchor="middle"
          class="axis-label"
        >
          {item.label.split(" ")[0]}
        </text>
      {/each}
    </svg>
  </div>

  <!-- Sequential Heatmap -->
  <div class="preview-card heatmap-card" style="background-color: {cardColor};">
    <div class="chart-title" style="color: {fgPrimary};">
      Activity <span class="font-normal" style="opacity: 0.6;"
        >(Sequential)</span
      >
    </div>
    <svg
      class="heatmap-svg"
      viewBox="0 0 {heatmapWidth} {heatmapHeight}"
      preserveAspectRatio="xMidYMid meet"
    >
      {#each seqCols as col, j}
        <text
          x={heatmapPadding.left + j * seqCellWidth + seqCellWidth / 2}
          y={heatmapPadding.top - 6}
          text-anchor="middle"
          class="axis-label"
        >
          {col}
        </text>
      {/each}

      {#each seqData as row, i}
        <text
          x={heatmapPadding.left - 4}
          y={heatmapPadding.top + i * seqCellHeight + seqCellHeight / 2}
          text-anchor="end"
          dominant-baseline="middle"
          class="axis-label"
        >
          {seqRows[i]}
        </text>

        {#each row as value, j}
          <rect
            x={heatmapPadding.left + j * seqCellWidth + 1}
            y={heatmapPadding.top + i * seqCellHeight + 1}
            width={seqCellWidth - 2}
            height={seqCellHeight - 2}
            fill={sequentialColors[value - 1] || primaryColor}
            rx="2"
          />
        {/each}
      {/each}
    </svg>
  </div>

  <!-- Diverging Heatmap -->
  <div class="preview-card heatmap-card" style="background-color: {cardColor};">
    <div class="chart-title" style="color: {fgPrimary};">
      YoY Change <span class="font-normal" style="opacity: 0.6;"
        >(Diverging)</span
      >
    </div>
    <svg
      class="heatmap-svg"
      viewBox="0 0 {heatmapWidth} {heatmapHeight}"
      preserveAspectRatio="xMidYMid meet"
    >
      {#each divCols as col, j}
        <text
          x={heatmapPadding.left + j * divCellWidth + divCellWidth / 2}
          y={heatmapPadding.top - 6}
          text-anchor="middle"
          class="axis-label"
        >
          {col}
        </text>
      {/each}

      {#each divData as row, i}
        <text
          x={heatmapPadding.left - 4}
          y={heatmapPadding.top + i * divCellHeight + divCellHeight / 2}
          text-anchor="end"
          dominant-baseline="middle"
          class="axis-label"
        >
          {divRows[i]}
        </text>

        {#each row as value, j}
          <rect
            x={heatmapPadding.left + j * divCellWidth + 1}
            y={heatmapPadding.top + i * divCellHeight + 1}
            width={divCellWidth - 2}
            height={divCellHeight - 2}
            fill={divergingColors[divValueToColorIndex(value)] || primaryColor}
            rx="2"
          />
        {/each}
      {/each}
    </svg>
  </div>
</div>

<style lang="postcss">
  .preview-card {
    @apply rounded border p-3 shadow-sm border-gray-200;
    min-height: 140px;
  }

  :global(.dark) .preview-card {
    @apply border-gray-600;
  }

  .kpi-card {
    @apply flex flex-col items-center justify-center gap-2;
  }

  .kpi-data-wrapper {
    @apply flex flex-col items-center w-full;
  }

  .kpi-header {
    @apply text-sm font-medium truncate w-full text-center;
  }

  .kpi-value {
    @apply text-xl font-medium;
  }

  .kpi-comparison {
    @apply flex items-center gap-2 text-sm;
  }

  .kpi-prev,
  .kpi-percent {
    @apply font-medium;
  }

  .kpi-delta {
    @apply font-medium;
  }

  .kpi-label {
    @apply text-sm;
  }

  .kpi-sparkline-wrapper {
    @apply w-full flex flex-col mt-1;
  }

  .kpi-sparkline-svg {
    @apply w-full overflow-visible;
    height: 32px;
  }

  .kpi-dates {
    @apply flex justify-between text-xs mt-0.5;
  }

  .chart-title {
    @apply text-sm font-medium mb-2;
  }

  .leaderboard-card {
    @apply flex flex-col;
  }

  .leaderboard-table {
    @apply w-full p-0 m-0 border-spacing-0 border-collapse table-fixed;
  }

  .leaderboard-row {
    @apply cursor-pointer;
    height: 22px;
  }

  .leaderboard-row:hover td {
    background: linear-gradient(
      to right,
      var(--color-theme-200) var(--bar-length, 0px),
      var(--color-gray-100) var(--bar-length, 0px)
    ) !important;
  }

  .leaderboard-name {
    @apply text-xs text-left px-2 truncate;
    width: 90px;
  }

  .leaderboard-value {
    @apply text-xs text-right px-2;
  }

  .bar-chart-card {
    @apply flex flex-col;
  }

  .bar-chart-svg {
    @apply w-full flex-1;
    min-height: 90px;
  }

  .heatmap-card {
    @apply flex flex-col;
  }

  .heatmap-svg {
    @apply w-full flex-1;
    min-height: 80px;
  }

  .axis-label {
    font-size: 9px;
    @apply fill-gray-500;
  }

  :global(.dark) .axis-label {
    @apply fill-gray-400;
  }

  .grid-line {
    stroke-width: 0.5;
    @apply stroke-gray-200;
  }

  :global(.dark) .grid-line {
    @apply stroke-gray-600;
  }
</style>
