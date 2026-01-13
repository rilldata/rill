<script lang="ts">
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";

  export let apiUrl: string;
</script>

<div class="flex flex-col gap-y-4">
  <p class="text-sm text-muted-foreground">
    Add this URL to your AI client's MCP server settings:
  </p>
  <CLICommandDisplay command={apiUrl} />
  <p class="text-xs text-muted-foreground">
    The OAuth flow will start automatically in your browser.
    <a
      href="https://docs.rilldata.com/explore/mcp#connect-using-oauth-recommended"
      target="_blank"
      rel="noopener"
    >
      Learn more
    </a>
  </p>
</div>
