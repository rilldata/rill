<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ClaudeIcon from "@rilldata/web-common/components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "@rilldata/web-common/components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "@rilldata/web-common/components/icons/connectors/OpenAIIcon.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { get } from "svelte/store";
  import { getConnectClientContext } from "./connect-client-context";

  const { adminServer } = featureFlags;
  // Only Rill Cloud wires the connect flow (see ConnectClientProvider); elsewhere
  // the hero renders nothing.
  const connectClient = get(adminServer) ? getConnectClientContext() : null;
</script>

{#if connectClient}
  <div class="connect-hero">
    <div class="connect-hero-title">
      Query your Rill metrics from your favorite AI client.
    </div>
    <div class="connect-hero-subtitle">Or directly ask questions in Rill.</div>
    <div class="connect-hero-buttons">
      <Button type="secondary" onClick={connectClient.open}>
        <ClaudeIcon size="16px" />
        Claude
      </Button>
      <Button type="secondary" onClick={connectClient.open}>
        <OpenAIIcon size="16px" />
        ChatGPT
      </Button>
      <Button type="secondary" onClick={connectClient.open}>
        <GeminiIcon size="16px" />
        Gemini
      </Button>
      <Button type="text" onClick={connectClient.open}>Other client</Button>
    </div>
  </div>
{/if}

<style lang="postcss">
  .connect-hero {
    @apply mt-8 w-full max-w-md;
    @apply flex flex-col items-center gap-2;
    @apply mx-auto p-5;
    @apply rounded-lg border border-border bg-surface-base;
  }

  .connect-hero-title {
    @apply text-sm font-semibold text-fg-primary text-center;
  }

  .connect-hero-subtitle {
    @apply mb-2 text-xs text-fg-secondary text-center;
  }

  .connect-hero-buttons {
    @apply flex flex-wrap items-center justify-center gap-2;
  }
</style>
