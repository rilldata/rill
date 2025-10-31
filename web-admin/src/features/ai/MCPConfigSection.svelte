<script lang="ts">
  import { CANONICAL_ADMIN_API_URL } from "@rilldata/web-admin/client/http-client";
  import CodeBlock from "@rilldata/web-common/components/code-block/CodeBlock.svelte";

  export let organization: string;
  export let project: string;
  export let isPublic: boolean;
  export let issuedToken: string | null = null;

  // Construct the API URL for the MCP server
  $: apiUrl = `${CANONICAL_ADMIN_API_URL}/v1/orgs/${organization}/projects/${project}/runtime/mcp`;

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
        "AUTH_HEADER": "Bearer ${issuedToken ? issuedToken : "<Rill personal access token>"}"
      }
    }
  }
}`;
</script>

<div class="flex flex-col gap-y-3">
  <h3 class="text-sm font-bold uppercase tracking-wide text-gray-900">
    Configure your MCP client
  </h3>
  <p class="text-sm text-gray-600">
    Use the below snippet to configure your AI client.
  </p>
  <CodeBlock code={isPublic ? publicConfig : privateConfig} language="json" />
</div>
