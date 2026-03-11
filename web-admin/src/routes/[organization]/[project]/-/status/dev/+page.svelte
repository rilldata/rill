<script lang="ts">
  import { page } from "$app/stores";
  import DevEnvironmentsSection from "@rilldata/web-admin/features/projects/status/dev-environments/DevEnvironmentsSection.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <DevEnvironmentsSection {organization} {project} />
</div>
