<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import OrganizationHibernating from "@rilldata/web-admin/features/organizations/hibernating/OrganizationHibernating.svelte";
  import { areAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
  import {
    createAdminServiceCreateProject,
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import OrganizationHero from "../../features/organizations/OrganizationHero.svelte";
  import ProjectCards from "../../features/projects/ProjectCards.svelte";
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

  const createProjectMutation = createAdminServiceCreateProject();
  async function createProject() {
    const createProjectResp = await $createProjectMutation.mutateAsync({
      org: orgName,
      data: {
        project: `${orgName}_project`,
        generateManagedGit: true,
        prodSlots: "4",
      },
    });
    const frontendUrl = createProjectResp.project?.frontendUrl;
    if (!frontendUrl) return;
    await goto(`${frontendUrl}/-/welcome`);
  }
  $: ({ isLoading } = $createProjectMutation);
</script>

<svelte:head>
  <title>{title} overview - Rill</title>
</svelte:head>

<ContentContainer showTitle={false} maxWidth={1300}>
  <Button type="primary" onClick={createProject} loading={isLoading}>
    Create new
  </Button>
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
    {:else if $allProjectsHibernating.data}
      <OrganizationHibernating
        organization={orgName}
        {organizationPermissions}
      />
    {:else}
      <div class="flex flex-col gap-y-8">
        <OrganizationHero {title} />
        <ProjectCards organization={orgName} />
      </div>
    {/if}
  {/if}
</ContentContainer>
