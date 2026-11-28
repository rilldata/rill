<script context="module" lang="ts">
  const zigZag =
    "M" +
    Array.from({ length: 7 })
      .map((_, i) => {
        return `${15 - 4 * (i % 2)} ${1.7 * (i * 2)}`;
      })
      .join(" L");
</script>

<!-- Drawn for a 22px row and stretched to the row's height, which follows the
     text size preset of a canvas leaderboard. -->
<svg
  class="absolute right-0 top-0 z-50"
  width="15"
  viewBox="0 0 15 22"
  preserveAspectRatio="none"
  style:height="calc(22px * var(--text-scale, 1))"
>
  <path d={zigZag} class="fill-surface" />
</svg>
