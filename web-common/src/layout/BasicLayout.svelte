<!-- @component
Combines a Navigation element with a slot for the main content of the page.
BasicLayout is the backbone of the Rill application.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import TooltipRenderer from "./TooltipRenderer.svelte";

  import Navigation from "./navigation/Navigation.svelte";

  $: showNavigation = $page.route.id !== "/(application)/welcome";
</script>

<TooltipRenderer>
  <main class="flex overflow-hidden h-screen w-screen">
    {#if showNavigation}
      <Navigation />
    {/if}
    <slot />
  </main>
</TooltipRenderer>
