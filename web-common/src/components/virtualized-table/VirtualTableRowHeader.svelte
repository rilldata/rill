<script lang="ts">
  export let index: number | string;
</script>

<span>
  {index}
</span>

<style lang="postcss">
  span {
    @apply bg-gray-100 size-full flex;
    @apply items-center justify-center font-bold p-0;
  }
</style>
