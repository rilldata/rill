<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { KeyIcon } from "lucide-svelte";
  import { LOCAL_STORAGE_KEY } from "../constants";

  export let connection: string;
  export let showTooltip = false;

  function getLastUsedConnection() {
    return localStorage.getItem(LOCAL_STORAGE_KEY);
  }

  $: lastUsedConnection = getLastUsedConnection();
  $: showTooltip = lastUsedConnection === connection;
</script>

{#if showTooltip}
  <Tooltip.Root portal="body" open={showTooltip}>
    <Tooltip.Trigger>
      <slot />
    </Tooltip.Trigger>
    <Tooltip.Content side="right" sideOffset={12}>
      <div class="flex items-center gap-x-1">
        <KeyIcon size={12} />
        <span>Last used</span>
      </div>
    </Tooltip.Content>
  </Tooltip.Root>
{:else}
  <slot />
{/if}
