<script lang="ts">
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { workspaceRoutePrefix } from "@rilldata/web-common/features/workspaces/workspace-routing";
  import type { LayoutData } from "../$types";

  export let data: LayoutData;

  // web-local uses top-level routes; no prefix needed
  $workspaceRoutePrefix = "";
</script>

<div class="flex size-full overflow-hidden">
  {#if data.initialized}
    <Navigation />
  {/if}

  <section class="size-full overflow-hidden">
    <slot />
  </section>
</div>
