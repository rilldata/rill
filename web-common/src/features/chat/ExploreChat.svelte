<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";
  import {
    getStateManagers,
    DEFAULT_STORE_KEY,
  } from "../dashboards/state-managers/state-managers";
  import { getContext } from "svelte";
  import type { StateManagers } from "../dashboards/state-managers/state-managers";

  const { dashboardChat } = featureFlags;

  // Get dashboard state managers if available (when used in explore dashboard context)
  let stateManagers: StateManagers | undefined;
  try {
    stateManagers = getContext(DEFAULT_STORE_KEY);
  } catch (e) {
    // No dashboard context available - this is fine for generic chat
    stateManagers = undefined;
  }
</script>

{#if $dashboardChat && $chatOpen}
  <SidebarChat {stateManagers} />
{/if}
