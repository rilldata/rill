<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { getRepoNameFromGithubUrl } from "@rilldata/web-common/features/project/deploy/github-utils";
  import ProjectSelector from "@rilldata/web-common/features/project/deploy/ProjectSelector.svelte";
  import ProjectSelectorItem from "@rilldata/web-common/features/project/deploy/ProjectSelectorItem.svelte";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";
  import Rocket from "svelte-radix/Rocket.svelte";

  export let open = false;
  export let matchingProjects: Project[];

  let selectedProject: Project | undefined =
    matchingProjects.length === 1 ? matchingProjects[0] : undefined;

  $: githubUrl = selectedProject?.githubUrl ?? "";
  $: repoName = getRepoNameFromGithubUrl(githubUrl);
  $: subpath = selectedProject?.subpath;

  $: selfManagedGit = githubUrl && !selectedProject?.managedGitId;
  $: enableUpdate = selectedProject && !selfManagedGit;

  $: deployUrl = selectedProject
    ? `/deploy/redeploy?org=${selectedProject.orgName}&project_id=${selectedProject.id}`
    : "";
</script>

<Popover.Root bind:open>
  <Popover.Trigger asChild let:builder>
    <Button type="primary" builders={[builder]}>
      <Rocket size="16px" />

      Deploy
    </Button>
  </Popover.Trigger>
  <Popover.Content align="start" class="w-[420px] flex flex-col gap-y-2">
    <div class="text-base font-medium">Update</div>
    <div class="text-sm text-slate-500">Push local changes to Rill Cloud?</div>

    {#if matchingProjects.length === 1 && selectedProject}
      <div class="border rounded-sm border-gray-300">
        <ProjectSelectorItem project={selectedProject} />
      </div>
    {:else}
      <ProjectSelector bind:selectedProject projects={matchingProjects} />
    {/if}

    {#if selfManagedGit}
      <div>
        This project has already been connected to a GitHub repo. Please push
        changes directly to GitHub and the project in Rill Cloud will
        automatically be updated.
        <a
          href="https://docs.rilldata.com/deploy/deploy-dashboard/github-101"
          target="_blank"
        >
          Learn more ->
        </a>
      </div>
      <div class="w-fit mx-auto">
        <div class="flex flex-row gap-x-1 items-center">
          <Github className="w-4 h-4" />
          <a
            href={githubUrl}
            class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
            target="_blank"
            rel="noreferrer noopener"
          >
            {repoName}
          </a>
        </div>
        {#if subpath}
          <div class="flex items-center">
            <span class="font-mono">subpath</span>
            <span class="text-gray-800">
              : /{subpath}
            </span>
          </div>
        {/if}
      </div>
    {/if}

    <div class="flex flex-row-reverse items-center">
      <Button
        type="primary"
        disabled={!enableUpdate}
        href={deployUrl}
        target="_blank"
        on:click={() => (open = false)}
      >
        Update
      </Button>
      <!-- TODO -->
      <!-- <Button type="secondary">Deploy to another project</Button>-->
    </div>
  </Popover.Content>
</Popover.Root>
