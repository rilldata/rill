<script lang="ts">
  import { page } from "$app/stores";
  import DeploymentsPage from "@rilldata/web-admin/features/branches/DeploymentsPage.svelte";
  import BranchesSection from "@rilldata/web-admin/features/branches/BranchesSection.svelte";

  let organization = $derived($page.params.organization);
  let project = $derived($page.params.project);
</script>

<div class="size-full min-w-0 flex flex-col gap-8">
  <DeploymentsPage {organization} {project} />
  <BranchesSection {organization} {project} />
</div>
