<script lang="ts">
  import { page } from "$app/stores";
  import InfoRowAi from "./InfoRowAi.svelte";
  import InfoRowDeployment from "./InfoRowDeployment.svelte";
  import InfoRowGithub from "./InfoRowGithub.svelte";
  import InfoRowOlap from "./InfoRowOlap.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<dl class="info-list">
  <InfoRowDeployment {organization} {project} />
  <InfoRowGithub {organization} {project} />
  <InfoRowOlap />
  <InfoRowAi />
</dl>

<style lang="postcss">
  .info-list {
    @apply m-0; /* Reset default dl margin */
    @apply border border-slate-200 rounded-sm;
  }
</style>
