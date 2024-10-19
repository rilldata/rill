<script lang="ts">
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";

  export let data;
</script>

<div class="flex size-full overflow-hidden">
  {#if data.initialized}
    <Navigation />
  {/if}

  <section class="size-full overflow-hidden">
    <slot />
  </section>
</div>
