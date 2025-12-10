<script lang="ts">
  import { page } from "$app/stores";
  import ConfigurationRowDeployment from "./ConfigurationRowDeployment.svelte";
  import ConfigurationRowGithub from "./ConfigurationRowGithub.svelte";
  import ConfigurationRowOlap from "./ConfigurationRowOlap.svelte";
  import ConfigurationRowAi from "./ConfigurationRowAi.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<section class="configuration-section">
  <div class="configuration-table">
    <ConfigurationRowDeployment {organization} {project} />
    <ConfigurationRowGithub {organization} {project} />
    <ConfigurationRowOlap />
    <ConfigurationRowAi />
  </div>
</section>

<style lang="postcss">
  .configuration-section {
    @apply flex flex-col;
  }

  .configuration-table {
    @apply border border-slate-200 rounded-sm overflow-hidden;
  }
</style>
