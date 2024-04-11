<!-- @component
Combines a Navigation element with a slot for the main content of the page.
BasicLayout is the backbone of the Rill application.
-->
<script lang="ts">
  import { page } from "$app/stores";

  import Navigation from "./navigation/Navigation.svelte";

  $: route = $page.route.id;

  $: hideNavigation =
    route === "/(application)/welcome" ||
    (route?.includes("/dashboard/") && !route?.includes("/edit"));
</script>

<main class="flex overflow-hidden h-screen w-screen">
  {#if !hideNavigation}
    <Navigation />
  {/if}
  <section class="size-full overflow-hidden">
    <slot />
  </section>
</main>
