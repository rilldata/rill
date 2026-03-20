<script lang="ts">
  import { page } from "$app/stores";
  import DeploymentsPage from "@rilldata/web-admin/features/projects/status/deployments/DeploymentsPage.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <DeploymentsPage {organization} {project} />
</div>
