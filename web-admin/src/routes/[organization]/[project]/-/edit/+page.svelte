<script lang="ts">
  import {
    createAdminServiceListDeployments,
    createAdminServiceCreateDeployment,
    createAdminServiceDeleteDeployment,
  } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { DateTime } from "luxon";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { createAdminServiceListOrganizationMemberUsers } from "@rilldata/web-admin/client";

  $: ({
    params: { organization, project },
  } = $page);

  const createDeploymentMutation = createAdminServiceCreateDeployment();
  const deleteDeploymentMutation = createAdminServiceDeleteDeployment();

  $: deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    { environment: "dev" },
  );

  $: deployments = $deploymentsQuery.data?.deployments || [];

  async function createDeployment() {
    await $createDeploymentMutation.mutateAsync({
      org: organization,
      project: project,
      data: { environment: "dev" },
    });

    await $deploymentsQuery.refetch();
  }

  async function deleteDeployment(deploymentId: string) {
    await $deleteDeploymentMutation.mutateAsync({ deploymentId });

    await $deploymentsQuery.refetch();
  }

  $: organizationMemberUsersQuery =
    createAdminServiceListOrganizationMemberUsers(organization);

  $: users = $organizationMemberUsersQuery.data?.members || [];
</script>

<button on:click={createDeployment}>Create New Edit Session</button>

<div class="p-4 flex flex-col gap-y-1 items-center w-full">
  {#each deployments as deployment (deployment.id)}
    {@const user = users.find((u) => u.userId === deployment.ownerUserId)}
    <div
      class="border p-4 m-2 rounded-md h-18 w-full max-w-6xl shadow-sm flex items-center justify-between"
    >
      <div>
        <a
          href="/{organization}/{project}/-/edit/{deployment.id}"
          class="text-md w-full truncate font-bold">{deployment.id}</a
        >
        <p>{deployment.environment}</p>
      </div>
      <p>{deployment.branch}</p>
      <p>
        {DateTime.fromISO(deployment.updatedOn).toLocaleString(
          DateTime.DATETIME_MED,
        )}
      </p>
      {#if user}
        <img
          src={user.userPhotoUrl}
          alt="User avatar"
          class="rounded-full size-7"
        />
        <p>{user.userName}</p>
      {/if}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          <ThreeDot size="20px" />
        </DropdownMenu.Trigger>
        <DropdownMenu.Content>
          <DropdownMenu.Item
            on:click={async () => {
              await deleteDeployment(deployment.id);
            }}
          >
            Delete
          </DropdownMenu.Item>
          <DropdownMenu.Item
            href="/{organization}/{project}/-/edit/{deployment.id}"
          >
            Edit
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </div>
  {/each}
</div>
