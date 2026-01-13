<script lang="ts">
  import CodeBlock from "@rilldata/web-common/components/code-block/CodeBlock.svelte";
  import PersonalAccessTokensSection from "../../personal-access-tokens/PersonalAccessTokensSection.svelte";

  export let apiUrl: string;
  export let isPublic: boolean;

  let issuedToken: string | null = null;

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

<div class="flex flex-col gap-y-6 min-w-0">
  {#if !isPublic}
    <PersonalAccessTokensSection bind:issuedToken />
  {/if}

  <div class="flex flex-col gap-y-3 min-w-0">
    <h4 class="text-sm font-medium text-foreground">Configuration</h4>
    <p class="text-sm text-muted-foreground">
      Add this to your MCP client's configuration file.
      <a
        href="https://docs.rilldata.com/explore/mcp#manual-configuration-alternative-method"
        target="_blank"
        rel="noopener"
      >
        Learn more
      </a>
    </p>
    <div class="overflow-x-auto">
      <CodeBlock
        code={isPublic ? publicConfig : privateConfig}
        language="json"
      />
    </div>
  </div>
</div>
