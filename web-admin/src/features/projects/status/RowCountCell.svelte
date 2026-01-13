<script lang="ts">
  export let count: number | "loading" | "error" | undefined;

  $: formattedCount = formatCount(count);

  function formatCount(c: typeof count): string {
    if (c === undefined) return "-";
    if (c === "loading") return "Loading...";
    if (c === "error") return "Error";
    if (typeof c === "number") return c.toLocaleString();
    return "-";
  }
</script>

<div class="truncate text-right tabular-nums">
  <span
    class:text-gray-500={formattedCount === "-"}
    class:text-gray-400={formattedCount === "Loading..."}
    class:text-red-500={formattedCount === "Error"}
  >
    {formattedCount}
  </span>
</div>
