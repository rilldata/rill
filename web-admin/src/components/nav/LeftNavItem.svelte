<script lang="ts">
  export let label: string;
  export let selected: boolean;
  export let link: string;
</script>

<a class:selected href={link}>
  <span class="text-slate-900">{label}</span>
</a>

<style lang="postcss">
  a {
    @apply p-2 flex gap-x-1 items-center;
    @apply rounded-sm;
    @apply text-sm font-medium;
  }

  .selected {
    @apply bg-slate-100;
  }

  a:hover {
    @apply bg-slate-50;
  }
</style>
