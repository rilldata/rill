<script lang="ts">
  import ConnectClientProvider from "@rilldata/web-admin/features/ai/mcp/ConnectClientProvider.svelte";
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <ConnectClientProvider>
    <DashboardChat />
  </ConnectClientProvider>
</div>
