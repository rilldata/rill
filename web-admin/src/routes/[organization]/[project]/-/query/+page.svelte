<script lang="ts">
  import { page } from "$app/state";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import QueryWorkspace from "@rilldata/web-common/features/query/QueryWorkspace.svelte";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  const instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);
  let activeBranch = $derived(extractBranchFromPath(page.url.pathname));
  let cacheKey = $derived(
    `${organization}/${project}${activeBranch ? `@${activeBranch}` : ""}`,
  );
  let {
    data: instanceData,
    error: instanceError,
    isLoading,
  } = $derived($instanceQuery);
  let olapConnector = $derived(instanceData?.instance?.olapConnector ?? "");
  let sqlConsoleEnabled = $derived(
    instanceData?.instance?.featureFlags?.sqlConsole ?? false,
  );
</script>

<svelte:head>
  <title>SQL Console · {project} · Rill</title>
</svelte:head>

<div class="size-full min-h-0 overflow-hidden">
  {#if isLoading}
    <div
      class="flex size-full items-center justify-center gap-x-2 text-sm text-fg-secondary"
    >
      <Spinner size="20px" status={EntityStatus.Running} />
      Loading SQL Console
    </div>
  {:else if instanceError}
    <ErrorPage
      statusCode={500}
      header="Unable to load SQL Console"
      body={instanceError.message}
    />
  {:else if !sqlConsoleEnabled}
    <ErrorPage
      statusCode={404}
      header="Page not found"
      body="SQL Console is not enabled for this project."
    />
  {:else if olapConnector}
    <QueryWorkspace {cacheKey} connector={olapConnector} />
  {:else}
    <div
      class="flex size-full items-center justify-center text-sm text-fg-secondary"
    >
      No project OLAP connector is available.
    </div>
  {/if}
</div>
