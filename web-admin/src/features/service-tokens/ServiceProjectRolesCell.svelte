<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetService } from "@rilldata/web-admin/client";

  export let serviceName: string;
  export let hasProjectRoles: boolean;

  $: organization = $page.params.organization;
  $: serviceQuery = createAdminServiceGetService(organization, serviceName, {
    query: { enabled: hasProjectRoles },
  });
  $: memberships = $serviceQuery.data?.projectMemberships ?? [];
</script>

{#if !hasProjectRoles}
  <span class="text-fg-tertiary">-</span>
{:else if $serviceQuery.isPending}
  <span class="text-fg-tertiary">...</span>
{:else if memberships.length > 0}
  <div class="flex flex-col gap-y-0.5">
    {#each memberships as pm}
      <span class="text-xs">
        {pm.projectName}
        <span class="text-fg-tertiary">({pm.projectRoleName})</span>
      </span>
    {/each}
  </div>
{:else}
  <span class="text-fg-tertiary">-</span>
{/if}
