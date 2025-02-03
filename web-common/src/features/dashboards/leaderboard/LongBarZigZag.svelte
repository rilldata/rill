<script context="module" lang="ts">
  const zigZag =
    "M" +
    Array.from({ length: 7 })
      .map((_, i) => {
        return `${15 - 4 * (i % 2)} ${1.7 * (i * 2)}`;
      })
      .join(" L");
</script>

<svg class="absolute right-0 top-0 z-50" width="15" height="22">
  <path d={zigZag} class="fill-surface" />
</svg>
