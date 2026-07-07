<script lang="ts">
  import { CANONICAL_ADMIN_API_URL } from "@rilldata/web-admin/client/http-client";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import ClaudeIcon from "@rilldata/web-common/components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "@rilldata/web-common/components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "@rilldata/web-common/components/icons/connectors/OpenAIIcon.svelte";
  import {
    Tabs,
    UnderlineTabsList,
    UnderlineTabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import type { ConnectClientProvider } from "@rilldata/web-common/features/chat/connect/connect-client-context";
  import ManualSection from "./ManualSection.svelte";
  import OAuthSection from "./OAuthSection.svelte";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let isPublic: boolean;
  /** When set, brands the dialog header for the chosen client. */
  export let provider: ConnectClientProvider | null = null;

  let activeTab = "oauth";

  const PROVIDER_META = {
    claude: { label: "Claude", icon: ClaudeIcon },
    openai: { label: "ChatGPT", icon: OpenAIIcon },
    gemini: { label: "Gemini", icon: GeminiIcon },
  } as const;

  $: providerMeta =
    provider && provider !== "other" ? PROVIDER_META[provider] : null;

  $: apiUrl = `${CANONICAL_ADMIN_API_URL}/v1/orgs/${organization}/projects/${project}/runtime/mcp`;
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-3xl overflow-x-hidden">
    <Dialog.Header>
      <Dialog.Title>
        <span class="flex items-center gap-2">
          {#if providerMeta}
            <svelte:component this={providerMeta.icon} size="20px" />
            Connect {providerMeta.label}
          {:else}
            Connect your own AI client
          {/if}
        </span>
      </Dialog.Title>
      <Dialog.Description>
        Ask questions of your Rill project using natural language in
        {providerMeta ? ` ${providerMeta.label}` : " any AI client"} and other tools
        that support the Model Context Protocol (MCP).
      </Dialog.Description>
    </Dialog.Header>

    <Tabs
      value={activeTab}
      onValueChange={(value) => (activeTab = value)}
      class="mt-2 min-w-0"
    >
      <UnderlineTabsList>
        <UnderlineTabsTrigger value="oauth">
          OAuth
          <Tag color="gray" text="Recommended" height={18} class="ml-1.5" />
        </UnderlineTabsTrigger>
        <UnderlineTabsTrigger value="manual">Manual</UnderlineTabsTrigger>
      </UnderlineTabsList>

      <div class="pt-4 min-w-0 overflow-x-auto">
        {#if activeTab === "oauth"}
          <OAuthSection {apiUrl} />
        {:else}
          <ManualSection {apiUrl} {isPublic} />
        {/if}
      </div>
    </Tabs>
  </Dialog.Content>
</Dialog.Root>
