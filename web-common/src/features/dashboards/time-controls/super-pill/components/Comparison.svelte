<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
</script>

<button>
  <div class="gap-x-2 flex">
    <b>vs</b>

    <p class="line-clamp-1">previous year</p>
  </div>
  <span class="flex-none">
    <CaretDownIcon />
  </span>
</button>

<style lang="postcss">
  button {
    @apply gap-x-1;
  }
</style>
