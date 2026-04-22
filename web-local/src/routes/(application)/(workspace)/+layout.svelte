<script lang="ts">
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import type { LayoutData } from "../$types";

  export let data: LayoutData;
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
