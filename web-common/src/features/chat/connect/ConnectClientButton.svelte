<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { get } from "svelte/store";
  import { getConnectClientContext } from "./connect-client-context";

  // "full": labeled button for the expanded conversation-sidebar footer.
  // "square": icon-only button for the collapsed conversation-sidebar footer.
  // "icon": header icon button for the dashboard/developer sidebar chat.
  let { variant }: { variant: "full" | "square" | "icon" } = $props();

  const { adminServer } = featureFlags;
  // The MCPConnectDialog is wired only on Rill Cloud, where ConnectClientProvider
  // is an ancestor of every chat surface. Elsewhere there is nothing to open, so
  // the button renders nothing.
  const connectClient = get(adminServer) ? getConnectClientContext() : null;

  const label = "Connect your own client";
</script>

{#if connectClient}
  {#if variant === "full"}
    <div class="footer">
      <Button type="secondary" class="w-full" onClick={connectClient.open}>
        <APIIcon size="14px" className="!fill-current" />
        {label}
      </Button>
    </div>
  {:else if variant === "square"}
    <div class="collapsed-footer">
      <span title={label}>
        <Button type="secondary" square {label} onClick={connectClient.open}>
          <APIIcon size="14px" className="!fill-current" />
        </Button>
      </span>
    </div>
  {:else}
    <IconButton ariaLabel={label} bgGray onclick={connectClient.open}>
      <APIIcon size="16px" className="!fill-current" />
      <svelte:fragment slot="tooltip-content">{label}</svelte:fragment>
    </IconButton>
  {/if}
{/if}

<style lang="postcss">
  .footer {
    @apply shrink-0 p-3 border-t border-border mt-auto;
  }

  .collapsed-footer {
    @apply flex flex-col items-center p-3 mt-auto;
  }
</style>
