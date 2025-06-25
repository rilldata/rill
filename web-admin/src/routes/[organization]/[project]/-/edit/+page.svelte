<script lang="ts">
  import {
    createAdminServiceListDeployments,
    createAdminServiceCreateDeployment,
  } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";

  $: ({
    params: { organization, project },
  } = $page);

  const createDeploymentMutation = createAdminServiceCreateDeployment();

  $: deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
  );

  $: deployments = $deploymentsQuery.data?.deployments || [];

  async function createDeployment() {
    await $createDeploymentMutation.mutateAsync({
      organizationName: organization,
      projectName: project,
      data: { environment: "dev" },
    });

    $deploymentsQuery.refetch();
  }
</script>

<button on:click={createDeployment}>Create New Deployment</button>
<div class="flex flex-col gap-y-3">
  {#each deployments as deployment (deployment.id)}
    <div>
      <a href="/{organization}/{project}/-/edit/{deployment.id}"
        >{deployment.id}</a
      >
      <p>Status: {deployment.status}</p>
      <p>Created at: {deployment.createdOn}</p>
      <p>{JSON.stringify(deployment, null, 2)}</p>
    </div>
  {/each}
</div>
