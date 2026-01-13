<script lang="ts">
  export let count: number | undefined;

  $: formattedCount = formatCount(count);

  function formatCount(c: number | undefined): string {
    if (c === undefined) return "-";
    return c.toLocaleString();
  }
</script>

<div class="truncate text-right tabular-nums">
  <span class:text-gray-500={formattedCount === "-"}>
    {formattedCount}
  </span>
</div>
