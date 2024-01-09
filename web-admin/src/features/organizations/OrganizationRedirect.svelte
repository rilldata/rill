<script lang="ts">
  import { goto } from "$app/navigation";
  import { adminServiceListOrganizations } from "@rilldata/web-admin/client";
  import { onMount } from "svelte";
  import { LOCAL_STORAGE_ACTIVE_ORG_KEY } from "./activeOrg";

  let showWelcomeMessage = false;

  onMount(async () => {
    // Scenario 1: User has an activeOrg in localStorage
    const activeOrg = localStorage.getItem(LOCAL_STORAGE_ACTIVE_ORG_KEY);
    if (activeOrg) {
      await goto(`/${activeOrg}`);
      return;
    }

    const orgs = (await adminServiceListOrganizations()).organizations;

    // Scenario 2: User has no activeOrg in localStorage, but does belong to an org
    if (orgs.length > 0) {
      localStorage.setItem(LOCAL_STORAGE_ACTIVE_ORG_KEY, orgs[0].name);
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
