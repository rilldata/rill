<script lang="ts">
  import { page } from "$app/stores";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  $: showNavigation = $page.route.id !== "/(application)/welcome";
</script>

<main class="flex overflow-hidden h-screen w-screen">
  {#if showNavigation}
    <Navigation />
  {/if}
  <slot />
</main>
