<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ProjectDeploymentStatus from "@rilldata/web-admin/features/projects/status/ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "@rilldata/web-admin/features/projects/status/ProjectGithubConnection.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  // fetch resource status
  const resources = createRuntimeServiceListResources(
    $runtime.instanceId,
    // all kinds
    undefined,
    {
      query: {
        select: (data) => {
          // filter out the "ProjectParser" resource
          return data.resources.filter(
            (resource) =>
              resource.meta.name.kind !== "rill.runtime.v1.ProjectParser",
          );
        },
      },
    },
  );
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <div class="flex flex-col gap-y-12">
      <div class="flex gap-x-20 items-start">
        <ProjectGithubConnection {organization} {project} />
        <ProjectDeploymentStatus {organization} {project} />
      </div>
      <!-- Project resources -->
      {#if $resources.data}
        <ProjectResources resources={$resources.data} />
      {/if}
      <ProjectParseErrors />
    </div>
  </ContentContainer>
</VerticalScrollContainer>
