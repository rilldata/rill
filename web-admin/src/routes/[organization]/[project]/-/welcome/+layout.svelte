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
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";

  let { children }: { children: Snippet } = $props();

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);

  let activeBranch = $derived(extractBranchFromPath(page.url.pathname));
  let projectQuery = $derived(
    createAdminServiceGetProject(
      organization,
      project,
      activeBranch ? { branch: activeBranch } : undefined,
      {
        query: baseGetProjectQueryOptions,
      },
    ),
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

<div class="flex size-full overflow-hidden">
  <div class="scroll">
    <div class="wrapper column p-10 2xl:py-16">
      {#if isProjectAvailable && !!runtime.host}
        {#key runtimeKey}
          <FileAndResourceWatcher lifecycle="none">
            <div class="mx-auto my-auto">
              {@render children()}
            </div>
          </FileAndResourceWatcher>
        {/key}
      {/if}
    </div>
  </div>
</div>

<style lang="postcss">
  .scroll {
    @apply size-full overflow-x-hidden overflow-y-auto;
  }

  .wrapper {
    @apply w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .wrapper {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }

  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
