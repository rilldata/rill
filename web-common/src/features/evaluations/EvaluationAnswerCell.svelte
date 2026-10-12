<script lang="ts">
  let { value }: { value: unknown } = $props();
</script>

<span class="truncate">{value ? JSON.stringify(value) : "-"}</span>
