<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import {
    getCreateProjectRoute,
    getDeployRouteForLocalRepo,
  } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { createLocalServiceGitStatus } from "@rilldata/web-common/runtime-client/local-service.ts";
  import GithubRepoDetails from "@rilldata/web-common/features/project/GithubRepoDetails.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;
  const orgParam = data.org;

  const gitStatusQuery = createLocalServiceGitStatus();
  const deployUrl = getDeployRouteForLocalRepo(orgParam);
</script>

<div class="text-xl flex flex-col gap-y-4 items-center">
  <Github className="w-10 h-10" />
  <div>Connect to github</div>
</div>
<div class="text-base text-gray-500">
  We’ve detected a self-managed GitHub repo associated with this project:
</div>
<GithubRepoDetails
  gitRemote={$gitStatusQuery.data?.githubUrl ?? ""}
  branch={$gitStatusQuery.data?.branch ?? ""}
  subpath={$gitStatusQuery.data?.subpath ?? ""}
/>
<div class="text-base text-gray-500">
  In order to link Rill Cloud to this repo to sync project updates, you’ll need
  to authenticate and install Rill
</div>

<Button wide type="primary" href={$deployUrl}>Link to this repo</Button>

<Button
  wide
  type="ghost"
  href={getCreateProjectRoute(orgParam)}
  class="-mt-2 flex flex-row items-center"
>
  <span>I'll do it later</span>
  <Tooltip.Root>
    <Tooltip.Trigger>
      <InfoCircle />
    </Tooltip.Trigger>
    <Tooltip.Content side="right" sideOffset={8} class="w-80">
      Choosing not to link to this repo can create a scenario where your Rill
      Cloud project will be out of sync with your self-managed Git repo. You can
      also link to this repo from the Project Status page in Rill Cloud to
      prevent merge issues
    </Tooltip.Content>
  </Tooltip.Root>
</Button>
