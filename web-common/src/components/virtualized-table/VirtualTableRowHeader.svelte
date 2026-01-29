<script lang="ts">
  export let index: number | string;
</script>

<span>
  {index}
</span>

<style lang="postcss">
  span {
    @apply bg-surface-base size-full flex text-fg-secondary;
    @apply items-center justify-center font-bold p-0;
  }
</style>
