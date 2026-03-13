<script lang="ts">
  import { page } from "$app/stores";
  import DeploymentsSection from "@rilldata/web-admin/features/projects/status/dev-environments/DeploymentsSection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <DeploymentsSection {organization} {project} />
</div>
