<script lang="ts">
  export let height: number;
  export let top: number;
  export let left: number;
  export let orientation: "horizontal" | "vertical" = "vertical";
</script>

<div
  class="absolute bg-primary-300"
  style:height="{height}px"
  style:width={orientation === "horizontal" ? "100%" : "2px"}
  style:top="{top}px"
  style:left="{left}px"
/>
