<script lang="ts">
  import type { RuntimeServiceCompleteBody } from "../../runtime-client";
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  export let agent: string | undefined = undefined;
  export let additionalContext: Partial<RuntimeServiceCompleteBody> = {};

  const { dashboardChat, developerChat } = featureFlags;
</script>

{#if ($dashboardChat || $developerChat) && $chatOpen}
  <SidebarChat {agent} {additionalContext} />
{/if}
