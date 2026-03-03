<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";

  const { dashboardChat } = featureFlags;
</script>

{#if $dashboardChat}
  <ChatToggle />
{/if}
