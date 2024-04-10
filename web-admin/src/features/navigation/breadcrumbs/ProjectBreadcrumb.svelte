<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";
  import {
    createAdminServiceGetProject,
    createAdminServiceListProjectsForOrganization,
  } from "../../../client";
  import { isProjectPage, isReportPage } from "../nav-utils";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";

  $: orgName = $page.params.organization;
  $: projectName = $page.params.project;

  $: organization = createAdminServiceGetOrganization(orgName);
  $: project = createAdminServiceGetProject(orgName, projectName);
  $: projects = createAdminServiceListProjectsForOrganization(
    orgName,
    undefined,
    {
      query: {
        enabled: !!$organization.data?.organization,
      },
    },
  );

  $: onProjectPage = isProjectPage($page);
  $: onReportPage = isReportPage($page);
</script>

{#if $project.data?.project}
  <span class="text-gray-600">/</span>
  <BreadcrumbItem
    label={projectName}
    href={onReportPage
      ? `/${orgName}/${projectName}/-/reports`
      : `/${orgName}/${projectName}`}
    menuItems={$projects.data?.projects?.length > 1 &&
      $projects.data.projects.map((proj) => ({
        key: proj.name,
        main: proj.name,
      }))}
    menuKey={projectName}
    onSelectMenuItem={(project) => goto(`/${orgName}/${project}`)}
    isCurrentPage={onProjectPage}
  />
{/if}
