<script lang="ts">
  import { ArrowUp } from "lucide-svelte";

  export let size: string | number = "1em";
  export let className = "";
  export let backgroundClass = "fill-gray-200";
  export let arrowClass = "stroke-gray-500";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={className}
>
  <!-- Circular background -->
  <circle cx="12" cy="12" r="10" class={backgroundClass} />
  <!-- Lucide arrow-up icon -->
  <g transform="translate(12, 12)">
    <ArrowUp
      size={14}
      absoluteStrokeWidth={true}
      strokeWidth={2.5}
      x={-7}
      y={-7}
      class={arrowClass}
    />
  </g>
</svg>
