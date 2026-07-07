<script lang="ts" module>
  import type { Component } from "svelte";
  import ClaudeIcon from "@rilldata/web-common/components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "@rilldata/web-common/components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "@rilldata/web-common/components/icons/connectors/OpenAIIcon.svelte";
  import type { ConnectClientProvider } from "./connect-client-context";

  const PROVIDERS: Array<{
    value: ConnectClientProvider;
    label: string;
    icon: Component<{ size?: string }>;
  }> = [
    { value: "claude", label: "Claude", icon: ClaudeIcon },
    { value: "openai", label: "ChatGPT", icon: OpenAIIcon },
    { value: "gemini", label: "Gemini", icon: GeminiIcon },
  ];
</script>

<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { getConnectClientContext } from "./connect-client-context";

  const connectClient = getConnectClientContext();
</script>

<div class="flex flex-wrap items-center justify-center gap-2">
  {#each PROVIDERS as provider (provider.value)}
    <Button type="secondary" onClick={() => connectClient.open(provider.value)}>
      <provider.icon size="16px" />
      {provider.label}
    </Button>
  {/each}
  <Button type="text" onClick={() => connectClient.open("other")}>
    Other client
  </Button>
</div>
