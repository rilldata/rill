<script lang="ts">
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import type { LayoutData } from "../$types";

  export let data: LayoutData;
</script>

<div class="flex size-full overflow-hidden">
  {#if data.initialized}
    <Navigation />
  {/if}

  <section class="size-full overflow-hidden">
    <slot />
  </section>
</div>
