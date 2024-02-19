<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    adminServiceGetCurrentUser,
    adminServiceListOrganizations,
  } from "@rilldata/web-admin/client";
  import { onMount } from "svelte";
  import { getActiveOrgLocalStorageKey } from "./local-storage";

  let showWelcomeMessage = false;

  onMount(async () => {
    // Get the activeOrg local storage key for the current user
    const userId = (await adminServiceGetCurrentUser())?.user?.id;
    const activeOrgLocalStorageKey = getActiveOrgLocalStorageKey(userId);

    // Scenario 1: User has an activeOrg in localStorage
    const activeOrg = localStorage.getItem(activeOrgLocalStorageKey);
    if (activeOrg) {
      await goto(`/${activeOrg}`);
      return;
    }

    const orgs = (await adminServiceListOrganizations()).organizations;

    // Scenario 2: User has no activeOrg in localStorage, but does belong to an org
    if (orgs.length > 0) {
      localStorage.setItem(activeOrgLocalStorageKey, orgs[0].name);
      await goto(`/${orgs[0].name}`);
      return;
    }

    // Scenario 3: User does not belong to an org
    showWelcomeMessage = true;
  });
</script>

{#if showWelcomeMessage}
  <slot />
{/if}
