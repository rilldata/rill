<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import MCPConfigSection from "@rilldata/web-admin/features/ai/MCPConfigSection.svelte";
  import PersonalAccessTokensSection from "@rilldata/web-admin/features/personal-access-tokens/PersonalAccessTokensSection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: proj = createAdminServiceGetProject(organization, project);
  $: ({
    project: { public: isPublic },
  } = $proj.data);

  let issuedToken: string | null = null;
</script>

<ContentContainer maxWidth={1100}>
  <div class="flex flex-col gap-y-4 size-full">
    <div class="flex flex-col gap-y-2">
      <h1 class="text-2xl font-bold mt-4 mb-2">
        Integrate Rill with your AI client
      </h1>
      <p class="mb-2 text-gray-700">
        Ask questions of your Rill project using natural language in any AI
        client that supports the Model Context Protocol (MCP). <a
          href="https://docs.rilldata.com/explore/mcp"
          target="_blank"
          rel="noopener">Learn more about MCP in the Rill docs</a
        >
      </p>
    </div>
    {#if !isPublic}
      <PersonalAccessTokensSection bind:issuedToken />
    {/if}
    <MCPConfigSection {organization} {project} {isPublic} {issuedToken} />
  </div>
</ContentContainer>
