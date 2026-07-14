<script lang="ts">
  let { value }: { value: string | undefined } = $props();

  const formatted = $derived.by(() => {
    if (!value) return "—";
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  });
</script>

<span class="whitespace-nowrap">{formatted}</span>
