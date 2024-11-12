<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { showWelcomeToRillDialog } from "@rilldata/web-admin/features/billing/plans/utils";
  import WelcomeToRillCloudDialog from "@rilldata/web-admin/features/billing/plans/WelcomeToRillCloudDialog.svelte";
  import { getActiveOrgLocalStorageKey } from "@rilldata/web-admin/features/organizations/active-org/local-storage";

  const user = createAdminServiceGetCurrentUser();
  $: organization = $page.params.organization;

  $: if ($user.data?.user?.id) {
    // get active org key for the current user
    const activeOrgLocalStorageKey = getActiveOrgLocalStorageKey(
      $user.data?.user?.id,
    );
    // store the navigated org to the local storage
    localStorage.setItem(activeOrgLocalStorageKey, organization);
  }
</script>

<slot />

<WelcomeToRillCloudDialog bind:open={$showWelcomeToRillDialog} />
