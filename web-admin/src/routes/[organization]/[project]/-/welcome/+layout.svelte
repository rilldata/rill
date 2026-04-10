<script lang="ts">
  import { page } from "$app/state";
  import type { Snippet } from "svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import {
    createAdminServiceGetProject,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/project-query-options.ts";
  import { resolveRuntimeConnection } from "@rilldata/web-admin/features/projects/project-runtime.ts";

  let { children }: { children: Snippet } = $props();

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);

  let projectQuery = $derived(
    createAdminServiceGetProject(organization, project, undefined, {
      query: baseGetProjectQueryOptions,
    }),
  );

  let projectData = $derived($projectQuery.data);

  let deploymentStatus = $derived(projectData?.deployment?.status);
  let isProjectAvailable = $derived(
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING,
  );

  let runtime = $derived(
    resolveRuntimeConnection(projectData, undefined, false),
  );
  let runtimeKey = $derived(
    `${runtime.host ?? ""}__${runtime.instanceId ?? ""}`,
  );
</script>

{#if isProjectAvailable && !!runtime.host}
  {#key runtimeKey}
    <FileAndResourceWatcher host={runtime.host} instanceId={runtime.instanceId}>
      <div class="mx-auto my-auto">
        {@render children()}
      </div>
    </FileAndResourceWatcher>
  {/key}
{/if}
