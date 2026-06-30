<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import LeftNav from "@rilldata/web-admin/components/nav/LeftNav.svelte";

  const basePage = "/-/superuser";
  const baseRoute = "/-/superuser";

  const navItems = [
    { label: "Users", route: "" },
    { label: "Superusers", route: "/superusers" },
    { label: "Billing", route: "/billing" },
    { label: "Quotas", route: "/quotas" },
    { label: "Organizations", route: "/organizations" },
    { label: "Projects", route: "/projects" },
  ];
</script>

<svelte:head>
  <title>Superuser Console | Rill</title>
</svelte:head>

<ContentContainer title="Superuser Console" maxWidth={1100}>
  <div class="container flex-col md:flex-row">
    <LeftNav {basePage} {baseRoute} {navItems} minWidth="180px" />
    <div class="flex flex-col gap-y-6 w-full">
      <slot />
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .container {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }
</style>
