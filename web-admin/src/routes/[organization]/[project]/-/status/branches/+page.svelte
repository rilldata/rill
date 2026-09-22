<script lang="ts">
  import { page } from "$app/stores";
  import BranchesSection from "@rilldata/web-admin/features/branches/BranchesSection.svelte";
  import ProjectSlotsSection from "@rilldata/web-admin/features/branches/ProjectSlotsSection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="size-full min-w-0 flex flex-col gap-6">
  <BranchesSection {organization} {project} />
  <ProjectSlotsSection {organization} {project} />
</div>
