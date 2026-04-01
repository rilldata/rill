<script lang="ts">
  import { page } from "$app/stores";
  import { isBrandingRestricted } from "@rilldata/web-admin/features/billing/plans/utils";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/resource-table/ProjectResources.svelte";

  $: brandingRestricted = isBrandingRestricted(
    $page.data.organization?.billingPlanName ?? "",
  );
</script>

<div class="size-full min-h-[600px]">
  <ProjectResources {brandingRestricted} />
</div>
