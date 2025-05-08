<script lang="ts">
  import { page } from "$app/stores";
  import LeftNavItem from "@rilldata/web-admin/components/nav/LeftNavItem.svelte";

  export let basePage: string;
  export let baseRoute: string;
  export let navItems: {
    label: string;
    route: string;
    hasPermission: boolean;
  }[];
  export let minWidth: string = "180px";
</script>

<div class="nav-items" style:min-width={minWidth}>
  {#each navItems as { label, route, hasPermission } (route)}
    {#if hasPermission}
      <LeftNavItem
        {label}
        link={`${basePage}${route}`}
        selected={$page.route.id === `${baseRoute}${route}`}
      />
    {/if}
  {/each}
</div>

<style lang="postcss">
  .nav-items {
    @apply flex flex-col gap-y-2;
  }
</style>
