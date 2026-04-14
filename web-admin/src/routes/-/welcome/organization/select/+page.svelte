<script lang="ts">
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { createAdminServiceListOrganizations } from "@rilldata/web-admin/client/index.ts";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control.ts";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo.ts";
  import { ChevronRightIcon, PlusIcon } from "lucide-svelte";
  import { InWelcomeFlowStore } from "@rilldata/web-admin/features/welcome/welcome-store.ts";

  const orgListQuery = createAdminServiceListOrganizations();
  $: orgs = $orgListQuery.data?.organizations ?? [];

  $: selectedTheme = $themeControl;

  function handleSelectOrg() {
    InWelcomeFlowStore.set(false);
  }
</script>

<div class="container">
  <RillLogoSquareNegative size="36px" />
  <div class="title">Choose an organization</div>

  <div class="content">
    <div class="orgs-list">
      {#each orgs as org (org.name)}
        {@const logoUrl = getThemedLogoUrl(selectedTheme, org)}
        <a class="link" href="/{org.name}" onclick={handleSelectOrg}>
          {#if logoUrl}
            <img src={logoUrl} alt="logo" class="h-8" />
          {:else}
            <Rill height="32" />
          {/if}
          <span class="grow">{org.displayName || org.name}</span>
          <ChevronRightIcon class="h-4" strokeWidth={1} />
        </a>
      {/each}
    </div>

    <a class="link" href="/-/welcome/organization/create">
      <PlusIcon class="h-4" />
      <span>Create a new organization</span>
    </a>
  </div>
</div>

<style lang="postcss">
  .container {
    @apply flex flex-col gap-4 mx-auto w-[486px];
  }

  .title {
    @apply text-2xl font-extrabold text-fg-accent text-center;
  }

  .content {
    @apply flex flex-col gap-8 pt-6;
  }

  .orgs-list {
    @apply flex flex-col gap-2;
  }

  .link {
    @apply flex flex-row w-full min-h-14 items-center gap-2 px-3 py-2 border rounded-sm;
    @apply bg-surface-base hover:bg-surface-hover border rounded-sm;
    @apply text-fg-primary text-sm font-medium;
  }
</style>
