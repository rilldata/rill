<script lang="ts">
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import type { LayoutData } from "../$types";
  import { page } from "$app/stores";

  export let data: LayoutData;

  $: ({
    params: { organization, project, deployment },
  } = $page);

  $: console.log($page);

  $: basePath = `/${organization}/${project}/-/edit/${deployment}`;

  $: console.log({ basePath });
</script>

<div class="flex size-full overflow-hidden">
  {#if data.initialized}
    <Navigation {basePath} />
  {/if}

  <section class="size-full overflow-hidden">
    <slot />
  </section>
</div>
