<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { createAdminServiceGetOrganization } from "../../client";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options";
  import { createQuery } from "@tanstack/svelte-query";
  import OrganizationHero from "../../features/organizations/OrganizationHero.svelte";
  import ProjectCards from "../../features/projects/ProjectCards.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { projectWelcomeEnabled } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";

  $: ({
    params: { organization: orgName },
  } = $page);

  $: org = createAdminServiceGetOrganization(orgName);
  $: projs = createQuery({
    ...listProjectsForOrgQueryOptions(orgName),
    enabled: !!$org.data?.organization,
  });

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
    {:else}
      <div class="flex flex-col gap-y-8">
        <OrganizationHero {title} />
        <ProjectCards organization={orgName} />
      </div>
    {/if}
  {/if}
</ContentContainer>
