<script lang="ts">
  import { page } from "$app/stores";
  import DeploymentsPage from "@rilldata/web-admin/features/projects/status/deployments/DeploymentsPage.svelte";

  let organization = $derived($page.params.organization);
  let project = $derived($page.params.project);
</script>

<DeploymentsPage {organization} {project} />
