<!-- PROJECT SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: basePage = `/${organization}/${project}/-/settings`;

  $: navItems = [
    {
      label: m.settings_nav_general(),
      route: "",
      hasPermission: true,
    },
    {
      label: m.settings_nav_env_vars(),
      route: "/environment-variables",
      hasPermission: true,
    },
    {
      label: m.settings_nav_public_urls(),
      route: "/public-urls",
      hasPermission: true,
    },
    {
      label: m.settings_nav_token_mgmt(),
      route: "/token-management",
      hasPermission: false,
    },
    {
      label: m.settings_nav_console(),
      route: "/console",
      hasPermission: false,
    },
  ];
</script>

<ContentContainer title={m.settings_project_page_title()} maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav
      {basePage}
      baseRoute="/[organization]/[project]/-/settings"
      {navItems}
      minWidth="180px"
    />
    <div class="flex flex-col gap-y-6 w-full min-w-0">
      {#if activeBranch}
        <Callout level="info">
          <span class="text-sm">
            {m.settings_branch_callout({ branch: activeBranch })}
          </span>
        </Callout>
      {/if}
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
