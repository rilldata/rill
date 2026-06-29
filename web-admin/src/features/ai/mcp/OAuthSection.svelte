<script lang="ts">
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let apiUrl: string;
</script>

<div class="flex flex-col gap-y-4">
  <p class="text-sm text-fg-secondary">
    {m.mcp_add_url()}
  </p>
  <CLICommandDisplay command={apiUrl} />
  <p class="text-xs text-fg-secondary">
    {m.mcp_oauth_auto()}
    <a
      href="https://docs.rilldata.com/guide/ai/mcp#connect-using-oauth-recommended"
      target="_blank"
      rel="noopener"
    >
      {m.mcp_learn_more()}
    </a>
  </p>
</div>
