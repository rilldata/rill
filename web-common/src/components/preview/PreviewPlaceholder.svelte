<script lang="ts">
  export let title: string;
  export let description: string;
</script>

<div
  class="flex flex-col items-center justify-center gap-y-3 rounded-md border border-dashed border-border bg-surface-subtle px-6 py-12 text-center"
>
  <slot name="icon" />
  <h3 class="text-sm font-semibold text-fg-primary">{title}</h3>
  <p class="text-sm text-fg-secondary max-w-md">{description}</p>
</div>
