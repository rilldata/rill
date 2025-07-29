<script lang="ts">
  export let size: number | string = 16;
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.2"
>
  <!-- Funnel shape with better spacing -->
  <rect x="2" y="2" width="20" height="2.5" rx="1" />
  <rect x="4" y="7" width="16" height="2.5" rx="1" />
  <rect x="6" y="12" width="12" height="2.5" rx="1" />
  <rect x="8" y="17" width="8" height="2.5" rx="1" />
</svg>
