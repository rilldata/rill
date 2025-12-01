<script lang="ts">
  import { CANONICAL_ADMIN_API_URL } from "@rilldata/web-admin/client/http-client";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import {
    Tabs,
    UnderlineTabsList,
    UnderlineTabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import ManualSection from "./ManualSection.svelte";
  import OAuthSection from "./OAuthSection.svelte";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let isPublic: boolean;

  let activeTab = "oauth";

  $: apiUrl = `${CANONICAL_ADMIN_API_URL}/v1/orgs/${organization}/projects/${project}/runtime/mcp`;
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-3xl overflow-x-hidden">
    <Dialog.Header>
      <Dialog.Title>Connect your own AI client</Dialog.Title>
      <Dialog.Description>
        Ask questions of your Rill project using natural language in any AI
        client that supports the Model Context Protocol (MCP).
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
