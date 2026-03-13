<!-- PROJECT SETTINGS -->

<script lang="ts">
  import { page } from "$app/stores";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: basePage = `/${organization}/${project}/-/settings`;

  const navItems = [
    {
      label: "General",
      route: "",
      hasPermission: true,
    },
    {
      label: "Environment Variables",
      route: "/environment-variables",
      hasPermission: true,
    },
    {
      label: "Public URLs",
      route: "/public-urls",
      hasPermission: true,
    },
    {
      label: "Token Management",
      route: "/token-management",
      hasPermission: false,
    },
    {
      label: "Console",
      route: "/console",
      hasPermission: false,
    },
  ];
</script>

<ContentContainer title="Project settings" maxWidth={1100}>
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
            These settings apply to the entire project, not just the
            <span class="font-mono">{activeBranch}</span> branch.
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
