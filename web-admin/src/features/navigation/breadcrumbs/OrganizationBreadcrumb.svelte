<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetOrganization,
    createAdminServiceListOrganizations,
  } from "../../../client";
  import { getActiveOrgLocalStorageKey } from "../../organizations/active-org/local-storage";
  import { isOrganizationPage } from "../nav-utils";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import OrganizationAvatar from "./OrganizationAvatar.svelte";

  const user = createAdminServiceGetCurrentUser();

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
    menuItems={$organizations.data?.organizations?.length > 1 &&
      $organizations.data.organizations.map((org) => ({
        key: org.name,
        main: org.name,
      }))}
    menuKey={orgName}
    onSelectMenuItem={onOrgChange}
    isCurrentPage={onOrganizationPage}
  >
    <OrganizationAvatar organization={orgName} slot="icon" />
  </BreadcrumbItem>
{/if}
