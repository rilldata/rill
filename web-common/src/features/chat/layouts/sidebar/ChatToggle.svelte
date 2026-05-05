<script lang="ts">
  import Button from "../../../../components/button/Button.svelte";
  import * as Tooltip from "../../../../components/tooltip-v2";
  import { chatOpen, sidebarActions } from "./sidebar-store";

  const isMac = window.navigator.userAgent.includes("Macintosh");
</script>

<svelte:window
  onkeydown={(e) => {
    if (e[isMac ? "metaKey" : "ctrlKey"] && e.key === "j") {
      e.preventDefault();
      sidebarActions.toggleChat();
    }
  }}
/>

<Tooltip.Root>
  <Tooltip.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        compact
        type="secondary"
        onClick={sidebarActions.toggleChat}
        active={$chatOpen}
      >
        AI
      </Button>
    {/snippet}
  </Tooltip.Trigger>
  <Tooltip.Content side="bottom">
    Open Rill AI {isMac ? "⌘" : "Ctrl"} + J
  </Tooltip.Content>
</Tooltip.Root>
