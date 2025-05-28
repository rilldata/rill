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
</script>

<ContentContainer maxWidth={1100}>
  <div class="flex flex-col gap-y-4 size-full">
    <MCPConfigSection {organization} {project} {isPublic} />
    <PersonalAccessTokensSection />
  </div>
</ContentContainer>
