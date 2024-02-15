<script lang="ts">
  /** the minimum container height */
  const MIN_CONTAINER_HEIGHT = "34px";
</script>

<section
  class="pl-2 grid gap-x-2 items-start"
  style:grid-template-columns="max-content auto"
  style:min-height={MIN_CONTAINER_HEIGHT}
>
  <slot />
</section>
