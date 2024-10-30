<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    adminServiceGetCurrentUser,
    adminServiceGetOrganizationNameForDomain,
    adminServiceListOrganizations,
  } from "@rilldata/web-admin/client";
  import { onMount } from "svelte";
  import { getActiveOrgLocalStorageKey } from "./local-storage";
  import { ADMIN_URL, CANONICAL_ADMIN_URL } from "../../../client/http-client";

  let showWelcomeMessage = false;

  onMount(async () => {
    // Scenario 1: If running on a custom domain, redirect to the org for the custom domain.
    if (ADMIN_URL !== CANONICAL_ADMIN_URL) {
      try {
        const res = await adminServiceGetOrganizationNameForDomain(
          window.location.hostname,
        );
        await goto(`/${res.name}`);
        return;
      } catch (e) {
        console.error("Failed to get organization for custom domain", e);
        // Fall back to the default behavior
      }
    }

    // Get the activeOrg local storage key for the current user
    const userId = (await adminServiceGetCurrentUser())?.user?.id;
    const activeOrgLocalStorageKey = getActiveOrgLocalStorageKey(userId);

    // Scenario 2: User has an activeOrg in localStorage
    const activeOrg = localStorage.getItem(activeOrgLocalStorageKey);
    if (activeOrg) {
      await goto(`/${activeOrg}`);
      return;
    }

    const orgs = (await adminServiceListOrganizations()).organizations;

    // Scenario 3: User has no activeOrg in localStorage, but does belong to an org
    if (orgs.length > 0) {
      await goto(`/${orgs[0].name}`);
      return;
    }

    // Scenario 4: User does not belong to an org
    showWelcomeMessage = true;
  });
</script>

{#if showWelcomeMessage}
  <slot />
{/if}
