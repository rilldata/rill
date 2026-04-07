<script lang="ts">
  export let value: string | undefined;

  function formatDate(v: string | undefined) {
    if (!v) return "—";
    return new Date(v).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }
</script>

<span class="whitespace-nowrap">{formatDate(value)}</span>
