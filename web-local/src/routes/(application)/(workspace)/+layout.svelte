<script lang="ts">
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { editRoutePrefix } from "@rilldata/web-common/layout/navigation/editor-routing";
  import type { LayoutData } from "../$types";

  export let data: LayoutData;

  // web-local uses top-level routes; no prefix needed
  $editRoutePrefix = "";
</script>

<div class="flex size-full overflow-hidden">
  {#if data.initialized}
    <Navigation />
  {/if}

  <section class="flex size-full overflow-hidden">
    <div class="flex-1 overflow-hidden">
      <slot />
    </div>
    <DeveloperChat />
  </section>
</div>
