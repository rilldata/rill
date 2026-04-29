<script lang="ts">
  import { page } from "$app/stores";
  import BranchesSection from "@rilldata/web-admin/features/branches/BranchesSection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="size-full min-w-0">
  <BranchesSection {organization} {project} />
</div>
