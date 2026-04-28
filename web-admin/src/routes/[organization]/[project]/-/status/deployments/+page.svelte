<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import DeploymentsPage from "@rilldata/web-admin/features/branches/DeploymentsPage.svelte";
  import BranchesSection from "@rilldata/web-admin/features/branches/BranchesSection.svelte";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";

  let organization = $derived($page.params.organization);
  let project = $derived($page.params.project);
  let activeBranch = $derived(extractBranchFromPath($page.url.pathname));

  $effect(() => {
    if (activeBranch) {
      void goto(
        `/${organization}/${project}${branchPathPrefix(activeBranch)}/-/status`,
        { replaceState: true },
      );
    }
  });
</script>

{#if !activeBranch}
  <div class="size-full min-w-0 flex flex-col gap-8">
    <DeploymentsPage {organization} {project} />
    <BranchesSection {organization} {project} />
  </div>
{/if}
