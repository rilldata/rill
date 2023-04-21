<script lang="ts">
  import { getProjectStatusStore } from "@rilldata/web-admin/components/projects/project-status-store";
  import type { ProjectStatusStore } from "@rilldata/web-admin/components/projects/project-status-store";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Timer from "@rilldata/web-common/components/icons/Timer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createAdminServiceGetProject } from "../../client";

  export let organization: string;
  export let project: string;

  const queryClient = useQueryClient();

  $: proj = createAdminServiceGetProject(organization, project);
  let projectStatusStore: ProjectStatusStore;
  $: projectStatusStore = getProjectStatusStore(
    organization,
    project,
    queryClient,
    proj
  );
</script>

{#if $projectStatusStore.queryRunning}
  <Spacer />
{:else if $projectStatusStore.pending}
  <Timer className="text-amber-600 hover:text-amber-500" />
{:else if $projectStatusStore.reconciling}
  <Spinner status={EntityStatus.Running} />
{:else if $projectStatusStore.ok}
  <CheckCircle className="text-blue-500 hover:text-blue-400" />
{:else if $projectStatusStore.errored}
  <CancelCircle className="text-red-500 hover:text-red-400" />
{/if}
