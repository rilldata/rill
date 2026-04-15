<script lang="ts">
  import { page } from "$app/stores";
  import BranchesSection from "@rilldata/web-admin/features/branches/BranchesSection.svelte";

  let organization = $derived($page.params.organization);
  let project = $derived($page.params.project);
</script>

<div class="size-full min-w-0">
  <BranchesSection {organization} {project} />
</div>
