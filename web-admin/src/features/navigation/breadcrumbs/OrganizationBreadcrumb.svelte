<script lang="ts">
    import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { useAlerts } from "@rilldata/web-admin/features/alerts/selectors";
  import { useValidDashboards } from "@rilldata/web-common/features/dashboards/selectors";
  import type {
    V1MetricsViewSpec,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetOrganization,
    createAdminServiceGetProject,
    createAdminServiceListOrganizations,
    createAdminServiceListProjectsForOrganization,
  } from "../../client";
  import { getActiveOrgLocalStorageKey } from "../organizations/active-org/local-storage";
  import { useReports } from "../scheduled-reports/selectors";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";
  import {
    isAlertPage,
    isDashboardPage,
    isOrganizationPage,
    isProjectPage,
    isReportPage,
  } from "./nav-utils";

  

  $: orgName = $page.params.organization;
  $: organization = createAdminServiceGetOrganization(orgName);
  $: organizations = createAdminServiceListOrganizations(
    { pageSize: 100 },
    {
      query: {
        enabled: !!$user.data?.user,
      },
    },
  );
  $: onOrganizationPage = isOrganizationPage($page);
  async function onOrgChange(org: string) {
    const activeOrgLocalStorageKey = getActiveOrgLocalStorageKey(
      $user.data?.user?.id,
    );
    localStorage.setItem(activeOrgLocalStorageKey, org);
    await goto(`/${org}`);
  }
</script>

{#if $organization.data?.organization}
      <BreadcrumbItem
        label={orgName}
        href={`/${orgName}`}
        menuOptions={$organizations.data?.organizations?.length > 1 &&
          $organizations.data.organizations.map((org) => ({
            key: org.name,
            main: org.name,
          }))}
        menuKey={orgName}
        onSelectMenuOption={onOrgChange}
        isCurrentPage={onOrganizationPage}
      >
        <OrganizationAvatar organization={orgName} slot="icon" />
      </BreadcrumbItem>
    {/if}