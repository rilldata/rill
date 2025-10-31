<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { getDeployRouteForProject } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { createLocalServiceGitStatus } from "@rilldata/web-common/runtime-client/local-service.ts";
  import GithubRepoDetails from "@rilldata/web-common/features/project/GithubRepoDetails.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  const orgParam = data.org;

  const gitStatusQuery = createLocalServiceGitStatus();
  const deployUrl = getDeployRouteForProject(orgParam);

  $: ({ isPending, data: statusData } = $gitStatusQuery);
  $: linkDisabled = isPending || !$deployUrl;
</script>

<div class="text-xl flex flex-col gap-y-4 items-center">
  <Github className="w-10 h-10" />
  <div>Connect to GitHub</div>
</div>
<div class="text-base text-gray-500">
  We’ve detected a self-managed GitHub repo associated with this project:
</div>
{#if isPending}
  <Spinner status={EntityStatus.Running} size="2rem" duration={725} />
{:else}
  <GithubRepoDetails
    gitRemote={statusData?.githubUrl ?? ""}
    branch={statusData?.branch ?? ""}
    subpath={statusData?.subpath ?? ""}
  />
{/if}
<div class="text-base text-gray-500">
  In order to link Rill Cloud to this repo to sync project updates, you’ll need
  to authenticate and install Rill
</div>

<Button
  wide
  type="primary"
  href={$deployUrl}
  loading={linkDisabled}
  disabled={linkDisabled}>Link to this repo</Button
>

<Button
  wide
  type="ghost"
  onClick={() => window.close()}
  class="-mt-2 flex flex-row items-center"
>
  Cancel
</Button>
