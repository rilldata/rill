<script lang="ts">
  export let label: string;
  export let selected: boolean;
  export let link: string;
</script>

<a class:selected href={link}>
  <span class="text-fg-primary">{label}</span>
</a>

<style lang="postcss">
  a {
    @apply p-2 flex gap-x-1 items-center;
    @apply rounded-sm;
    @apply text-sm font-medium;
  }

  .selected {
    @apply bg-surface-active;
  }

  a:hover {
    @apply bg-surface-hover;
  }
</style>
