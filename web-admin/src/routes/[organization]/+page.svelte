<script lang="ts">
  import { page } from "$app/stores";
  import OrganizationHibernating from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernating.svelte";
  import { areAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import ProjectCards from "../../features/projects/ProjectCards.svelte";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import OrganizationHero from "@rilldata/web-admin/features/organizations/OrganizationHero.svelte";
  import { projectWelcomeEnabled } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
  import { Button } from "@rilldata/web-common/components/button";

  export let data;
  $: ({ organizationPermissions } = data);
  $: ({
    params: { organization: orgName },
  } = $page);

  $: org = createAdminServiceGetOrganization(orgName);
  $: projs = createAdminServiceListProjectsForOrganization(orgName, undefined, {
    query: { enabled: !!$org.data?.organization },
  });
  $: allProjectsHibernating = areAllProjectsHibernating(orgName);

  $: title = $org.data?.organization?.displayName || orgName;
</script>

<svelte:head>
  <title>{title} overview - Rill</title>
</svelte:head>

<ContentContainer showTitle={false} maxWidth={1300}>
  {#if $org.data && $org.data.organization && $projs.data}
    {#if $projs.data.projects?.length === 0}
      <OrganizationHero {title} />
      <span>
        This organization has no projects yet. <a
          href="https://docs.rilldata.com/"
          target="_blank"
          rel="noreferrer noopener">See docs</a
        >
      </span>
      {#if projectWelcomeEnabled}
        <div class="w-fit">
          <Button type="primary" href="/{orgName}/-/create-project">
            Create new
          </Button>
        </div>
      {/if}
    {:else if $allProjectsHibernating.data}
      <OrganizationHibernating
        organization={orgName}
        {organizationPermissions}
      />
    {:else}
      <OrganizationHero {title} />
      <ProjectCards organization={orgName} />
    {/if}
  {/if}
</ContentContainer>
