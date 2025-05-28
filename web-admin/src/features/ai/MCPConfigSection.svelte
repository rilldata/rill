<script lang="ts">
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import CodeBlock from "@rilldata/web-common/components/code-block/CodeBlock.svelte";

  export let organization: string;
  export let project: string;
  export let isPublic: boolean;

  // Construct the API URL for the MCP server
  $: apiUrl = `${ADMIN_URL}/v1/organizations/${organization}/projects/${project}/runtime/mcp/sse`;

  // Config snippets with exact formatting
  $: publicConfig = `{
  "mcpServers": {
    "rill": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "${apiUrl}"
      ]
    }
  }
}`;

  $: privateConfig = `{
  "mcpServers": {
    "rill": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "${apiUrl}",
        "--header",
        "Authorization:\${AUTH_HEADER}"
      ],
      "env": {
        "AUTH_HEADER": "Bearer <Rill access token>"
      }
    }
  }
}`;
</script>

<div class="mb-8">
  <h2 class="text-xl font-semibold mb-2">Connect your MCP Client</h2>
  <p class="mb-4 text-gray-600">
    {#if isPublic}
      This project is <span class="font-medium">public</span>. Use the following
      configuration to connect your MCP-compatible client (e.g., Claude):
    {:else}
      This project is <span class="font-medium">private</span>. You will need a
      personal access token. Use the following configuration:
    {/if}
  </p>
  <CodeBlock code={isPublic ? publicConfig : privateConfig} language="json" />
</div>
