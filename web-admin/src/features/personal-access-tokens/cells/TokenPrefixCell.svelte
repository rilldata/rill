<script lang="ts">
  let { prefix }: { prefix: string | undefined } = $props();
</script>

<span
  class="rounded-sm bg-surface-subtle px-1.5 py-0.5 font-mono text-xs text-fg-secondary"
>
  {prefix}…
</span>
