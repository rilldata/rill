<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import APIIcon from "@rilldata/web-common/components/icons/APIIcon.svelte";
  import ClaudeIcon from "@rilldata/web-common/components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "@rilldata/web-common/components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "@rilldata/web-common/components/icons/connectors/OpenAIIcon.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { getConnectClientContext } from "./connect-client-context";

  // The MCPConnectDialog is wired only on Rill Cloud chat surfaces whose layout
  // sets the connect-client context. Elsewhere there is nothing to open, so the
  // popover renders nothing.
  const connectClient = getConnectClientContext();

  let isOpen = false;

  function openDialog() {
    isOpen = false;
    connectClient?.open();
  }
</script>

{#if connectClient}
  <Popover bind:open={isOpen}>
    <PopoverTrigger>
      <IconButton
        ariaLabel={m.chat_connect_client()}
        bgGray
        active={isOpen}
        disableTooltip={isOpen}
      >
        <APIIcon size="16px" className="!fill-current" />
        <svelte:fragment slot="tooltip-content">
          {m.chat_connect_client()}
        </svelte:fragment>
      </IconButton>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-[280px] p-4">
      <div class="flex flex-col gap-y-3">
        <div class="flex flex-col gap-y-1">
          <h3 class="text-sm font-medium text-fg-primary">
            {m.chat_connect_client()}
          </h3>
          <p class="text-xs text-fg-secondary">
            {m.chat_connect_client_description()}
          </p>
        </div>
        <div class="flex flex-col gap-y-1.5">
          <Button type="secondary" onClick={openDialog}>
            <ClaudeIcon size="16px" />
            <!-- i18n-ignore: standalone product name -->
            Claude
          </Button>
          <Button type="secondary" onClick={openDialog}>
            <OpenAIIcon size="16px" />
            <!-- i18n-ignore: standalone product name -->
            ChatGPT
          </Button>
          <Button type="secondary" onClick={openDialog}>
            <GeminiIcon size="16px" />
            <!-- i18n-ignore: standalone product name -->
            Gemini
          </Button>
          <Button type="text" onClick={openDialog}>
            {m.chat_connect_client_other()}
          </Button>
        </div>
      </div>
    </PopoverContent>
  </Popover>
{/if}
