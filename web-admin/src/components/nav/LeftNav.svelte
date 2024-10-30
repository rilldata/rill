<script lang="ts">
  import { page } from "$app/stores";
  import LeftNavItem from "@rilldata/web-admin/components/nav/LeftNavItem.svelte";

  export let basePage: string;
  export let baseRoute: string;
  export let navItems: {
    label: string;
    route: string;
  }[];
</script>

<div class="nav-items">
  {#each navItems as { label, route } (route)}
    <LeftNavItem
      {label}
      link={`${basePage}${route}`}
      selected={$page.route.id === `${baseRoute}${route}`}
    />
  {/each}
</div>

<style lang="postcss">
  .nav-items {
    @apply flex flex-col gap-y-2;
    @apply min-w-[150px];
  }
</style>
