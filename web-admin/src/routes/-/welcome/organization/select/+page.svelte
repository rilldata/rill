<script>
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { createAdminServiceListOrganizations } from "@rilldata/web-admin/client/index.ts";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control.ts";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo.ts";
  import { ChevronRightIcon } from "lucide-svelte";

  const orgListQuery = createAdminServiceListOrganizations();
  $: orgs = $orgListQuery.data?.organizations ?? [];

  $: selectedTheme = $themeControl;
</script>

<div class="container">
  <RillLogoSquareNegative size="36px" />
  <div class="title">Choose an organization</div>

  <div class="content">
    <div class="orgs-list">
      {#each orgs as org (org.name)}
        {@const logoUrl = getThemedLogoUrl(selectedTheme, org)}
        <a class="link" href="/{org.name}">
          {#if logoUrl}
            <img src={logoUrl} alt="logo" class="h-7" />
          {:else}
            <Rill />
          {/if}
          <span class="grow">{org.name}</span>
          <ChevronRightIcon class="h-4" strokeWidth={1} />
        </a>
      {/each}
    </div>

    <a class="link" href="/-/welcome/organization/create">
      Create a new organization
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
    @apply flex flex-row w-full min-h-11 items-center gap-2 px-3 py-2 border rounded-sm;
    @apply bg-surface-base hover:bg-surface-hover border rounded-sm;
    @apply text-fg-primary text-sm font-medium;
  }
</style>
