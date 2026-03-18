<script lang="ts">
  import SidebarChat from "@rilldata/web-common/features/chat/layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createQueryChatConfig } from "@rilldata/web-common/features/query/query-chat-config";
  import QueryWorkspace from "@rilldata/web-common/features/query/QueryWorkspace.svelte";
  import { page } from "$app/stores";

  const { dashboardChat } = featureFlags;
  const chatConfig = createQueryChatConfig();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex size-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <QueryWorkspace projectId="{organization}/{project}" />
  </div>
  {#if $dashboardChat && $chatOpen}
    <SidebarChat config={chatConfig} />
  {/if}
</div>
