<script lang="ts">
  import {
    getGitUrlFromRemote,
    getRepoNameFromGitRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";

  export let open = false;
  export let githubUrl: string;
  export let branch: string;
  export let subpath: string;
  export let onUseRill: () => void;
  export let onUseGithub: () => void;

  $: repoName = getRepoNameFromGitRemote(githubUrl);
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle
        >This project is already connected to a github repo</AlertDialogTitle
      >
      <AlertDialogDescription class="flex flex-col gap-y-4 pt-2">
        <div>
          Do you want to use this repo to deploy this project? Changes made
          directly to GitHub and the project in Rill Cloud will automatically be
          updated.
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
              href={getGitUrlFromRemote(githubUrl)}
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
          <div class="flex items-center">
            <span class="font-mono">branch</span>
            <span class="text-gray-800">
              : {branch}
            </span>
          </div>
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        onClick={() => {
          open = false;
          onUseRill();
        }}
        type="secondary"
      >
        Let rill manage it
      </Button>
      <Button
        onClick={() => {
          open = false;
          onUseGithub();
        }}
        type="secondary"
      >
        Use github
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
