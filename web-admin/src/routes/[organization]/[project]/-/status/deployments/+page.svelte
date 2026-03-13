<script lang="ts">
  import { page } from "$app/stores";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import DeploymentsSection from "@rilldata/web-admin/features/projects/status/dev-environments/DeploymentsSection.svelte";
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: activeBranch = extractBranchFromPath($page.url.pathname);
</script>

<div class="flex flex-col gap-y-4 size-full">
  {#if activeBranch}
    <Callout level="info">
      <span class="text-sm">
        Showing all deployments across all branches.
      </span>
    </Callout>
  {/if}
  <DeploymentsSection {organization} {project} />
</div>
