<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: basePath = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit/status`;

  $: navItems = [
    { label: "Overview", route: basePath },
    { label: "Resources", route: `${basePath}/resources` },
  ];
</script>

<ContentContainer title="Status" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <nav class="nav-items" style:min-width="180px">
      {#each navItems as { label, route } (route)}
        <a
          href={route}
          class="nav-item"
          class:selected={$page.url.pathname === route ||
            ($page.url.pathname.startsWith(route + "/") && route !== basePath)}
        >
          <span class="text-fg-primary">{label}</span>
        </a>
      {/each}
    </nav>
    <div class="flex flex-col gap-y-6 w-full overflow-hidden">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
  .nav-items {
    @apply flex flex-col gap-y-2;
  }
  .nav-item {
    @apply p-2 flex gap-x-1 items-center rounded-sm text-sm font-medium;
  }
  .selected {
    @apply bg-surface-active;
  }
  .nav-item:hover {
    @apply bg-surface-hover;
  }
</style>
