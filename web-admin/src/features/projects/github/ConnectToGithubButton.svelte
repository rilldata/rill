<script lang="ts">
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
  import { PopupWindow } from "@rilldata/web-common/lib/openPopupWindow";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";

  export let open = false;
  export let loading = false;
  export let isGithubConnected: boolean;
  export let onContinue: () => void;

  let githubRepoCreated = false;

  function openGithubRepoCreator() {
    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectCreateRepo,
      {
        is_fresh_connection: isGithubConnected,
      },
    );
    // we need a popup window so we cannot use href
    PopupWindow.open("https://github.com/new", "githubNewWindow");
  }

  function handleContinue() {
    open = false;
    githubRepoCreated = false;
    onContinue();
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild let:builder>
    <Button builders={[builder]} type="primary" class="w-fit mt-1" {loading}>
      <Github className="w-4 h-4" fill="white" />
      Connect to GitHub
    </Button>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <AlertDialogTitle>Connect to GitHub</AlertDialogTitle>
      </div>

      <AlertDialogDescription>
        <div class="mt-1">
          Before continuing, you need to create a repo in GitHub and grant Rill
          read and write access to it.
          <Button
            type="link"
            href="#"
            on:click={openGithubRepoCreator}
            forcedStyle="display:inline-block !important; padding: 0px !important; min-height:12px !important; height: 12px !important;"
          >
            <span class="text-sm">Create GitHub repo -></span>
          </Button>
        </div>
        <div class="flex gap-x-2 mt-4">
          <input
            type="checkbox"
            id="create-github-repo"
            bind:checked={githubRepoCreated}
          />
          <label for="create-github-repo" class="text-gray-800"
            >I’ve created a GitHub repo</label
          >
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-3">
      <Button type="secondary" on:click={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        on:click={handleContinue}
        disabled={!githubRepoCreated}
      >
        Continue
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
