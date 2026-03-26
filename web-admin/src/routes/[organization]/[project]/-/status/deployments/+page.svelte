<script lang="ts">
  import { page } from "$app/stores";
  import DeploymentsSection from "@rilldata/web-admin/features/projects/status/deployments/DeploymentsSection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="size-full min-w-0">
  <DeploymentsSection {organization} {project} />
</div>
