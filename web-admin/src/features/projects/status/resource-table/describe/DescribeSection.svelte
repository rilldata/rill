<script lang="ts">
  export let title = "";
</script>

<div class="border border-border rounded-md overflow-hidden">
  {#if title}
    <div class="px-3 py-2 border-b border-border bg-surface-subtle">
      <span
        class="text-xs font-semibold text-fg-secondary uppercase tracking-wide"
        >{title}</span
      >
    </div>
  {/if}
</div>
