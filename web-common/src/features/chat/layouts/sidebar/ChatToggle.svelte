<script lang="ts">
  import type { Readable } from "svelte/store";
  import Button from "../../../../components/button/Button.svelte";
  import * as Tooltip from "../../../../components/tooltip-v2";
  import type { ChatActions } from "./sidebar-store";

  export let open: Readable<boolean>;
  export let actions: ChatActions;

  const isMac = window.navigator.userAgent.includes("Macintosh");
</script>

<svelte:window
  onkeydown={(e) => {
    if (e[isMac ? "metaKey" : "ctrlKey"] && e.key === "j") {
      e.preventDefault();
      actions.toggleChat();
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
        onClick={actions.toggleChat}
        active={$open}
      >
        AI
      </Button>
    {/snippet}
  </Tooltip.Trigger>
  <Tooltip.Content side="bottom">
    Open Conversational AI {isMac ? "⌘" : "Ctrl"} + J
  </Tooltip.Content>
</Tooltip.Root>
