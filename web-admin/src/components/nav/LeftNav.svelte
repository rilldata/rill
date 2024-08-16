<script lang="ts">
  import { page } from "$app/stores";
  import LeftNavItem from "@rilldata/web-admin/components/nav/LeftNavItem.svelte";

  export let title: string;
  export let basePage: string;
  export let baseRoute: string;
  export let navItems: {
    label: string;
    route: string;
  }[];
</script>

<div class="nav-layout-container">
  <h3>{title}</h3>
  <div class="nav-container">
    <div class="nav-items">
      {#each navItems as { label, route } (route)}
        <LeftNavItem
          {label}
          link={`${basePage}/${route}`}
          selected={$page.route.id === `${baseRoute}/${route}`}
        />
      {/each}
    </div>
    <div><slot /></div>
  </div>
</div>

<style lang="postcss">
  .nav-layout-container {
    @apply px-32 py-10;
  }

  h3 {
    @apply text-2xl font-semibold;
  }

  .nav-container {
    @apply flex flex-row pt-6 gap-x-6;
  }

  .nav-items {
    @apply flex flex-col gap-y-1;
    @apply min-w-[150px];
  }
</style>
