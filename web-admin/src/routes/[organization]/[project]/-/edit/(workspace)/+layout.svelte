<script lang="ts">
  import { page } from "$app/stores";
  import { createProdRuntimeClient } from "@rilldata/web-admin/features/ai-feedback/prod-runtime-client";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";

  $: ({ organization, project } = $page.params);

  // The AI feedback badge counts open items on the production deployment too.
  $: prodRuntime = createProdRuntimeClient(organization, project);
</script>

<div class="flex flex-1 overflow-hidden">
  <Navigation
    showFooterLinks={false}
    extraFeedbackClient={$prodRuntime.client}
  />
  <section class="flex flex-1 overflow-hidden">
    <div class="flex-1 overflow-hidden">
      <slot />
    </div>
    <DeveloperChat />
  </section>
</div>
