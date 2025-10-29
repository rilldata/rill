<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import PersonalAccessTokensSection from "../personal-access-tokens/PersonalAccessTokensSection.svelte";
  import MCPConfigSection from "./MCPConfigSection.svelte";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let isPublic: boolean;

  let issuedToken: string | null = null;
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-3xl max-h-[80vh] overflow-y-auto">
    <Dialog.Header>
      <Dialog.Title>Connect your own AI client</Dialog.Title>
      <Dialog.Description>
        Ask questions of your Rill project using natural language in any AI
        client that supports the Model Context Protocol (MCP). <a
          href="https://docs.rilldata.com/explore/mcp"
          target="_blank"
          rel="noopener"
          class="text-primary-600 hover:text-primary-700 underline"
        >
          Learn more about MCP in the Rill docs
        </a>
      </Dialog.Description>
    </Dialog.Header>

    <div class="flex flex-col gap-y-6 min-w-0">
      {#if !isPublic}
        <div class="border-t pt-6">
          <PersonalAccessTokensSection bind:issuedToken />
        </div>
      {/if}
      <div class="border-t pt-6">
        <MCPConfigSection {organization} {project} {isPublic} {issuedToken} />
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>
