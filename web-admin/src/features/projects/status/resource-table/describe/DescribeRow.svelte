<script lang="ts">
  export let label: string;
  export let value: string | number | boolean | undefined = undefined;
  export let mono = true;
</script>

{#if value !== undefined && value !== null && value !== ""}
  <div class="flex items-baseline justify-between gap-x-4 min-h-[20px]">
    <span class="text-xs text-fg-secondary shrink-0">{label}</span>
    <span
      class="text-xs text-fg-primary text-right truncate"
      class:font-mono={mono}
      title={String(value)}
    >
      {value}
    </span>
  </div>
{/if}
