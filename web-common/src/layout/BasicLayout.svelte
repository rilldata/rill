<!-- @component
Combines a Navigation element with a slot for the main content of the page.
BasicLayout is the backbone of the Rill application.
-->
<script lang="ts">
  import { page } from "$app/stores";

  import Navigation from "./navigation/Navigation.svelte";

  $: showNavigation = $page.route.id !== "/(application)/welcome";
</script>

<main class="flex">
  {#if showNavigation}
    <Navigation />
  {/if}
  <slot />
</main>
