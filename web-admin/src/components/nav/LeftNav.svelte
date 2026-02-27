<script lang="ts">
  import { page } from "$app/stores";
  import LeftNavItem from "@rilldata/web-admin/components/nav/LeftNavItem.svelte";

  export let basePage: string;
  export let baseRoute: string;
  export let minWidth: string = "180px";
  export let navItems: {
    label: string;
    route: string;
    hasPermission?: boolean;
  }[];
</script>

<div class="nav-sidebar" style:min-width={minWidth}>
  <!-- if hasPermission is not provided, it will be undefined -->
  {#each navItems as { label, route, hasPermission = true } (route)}
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
  .nav-sidebar {
    @apply flex flex-col gap-y-2 shrink-0;
  }

  @media (min-width: 768px) {
    .nav-sidebar {
      position: sticky;
      top: 0;
      align-self: flex-start;
    }
  }
</style>
