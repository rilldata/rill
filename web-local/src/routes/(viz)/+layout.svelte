<script lang="ts">
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";

  $: showSidebar = !$previewModeStore;
</script>

<div class="flex size-full overflow-hidden">
  {#if showSidebar}
    <Navigation />
  {/if}

  <section class="flex size-full overflow-hidden">
    <div class="flex-1 overflow-hidden">
      <slot />
    </div>
  </section>
</div>
