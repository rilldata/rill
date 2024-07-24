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

  export let open = false;
  export let loading = false;
  export let onContinue: () => void;

  let githubRepoCreated = false;

  function openGithubRepoCreator() {
    // we need a popup window so we cannot use href
    window.open("https://github.com/new", "", "width=1024,height=600");
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
      Connect to Github
    </Button>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <div class="flex flex-row gap-x-2">
      <Github size="28px" />
      <div class="flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle>
            Connect this project to a Github repo
          </AlertDialogTitle>
          <AlertDialogDescription>
            <div>
              Before continuing, you need to create a repo in GitHub and grant
              Rill read and write access to it.
              <Button
                type="link"
                href="#"
                on:click={openGithubRepoCreator}
                forcedStyle="display:inline-block !important; padding: 0px !important; min-height:12px !important; height: 12px !important;"
              >
                Create GitHub repo ->
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
        <AlertDialogFooter class="mt-5">
          <Button type="secondary" on:click={() => (open = false)}>
            Cancel
          </Button>
          <Button
            type="primary"
            on:click={handleContinue}
            disabled={!githubRepoCreated}
          >
            Continue
          </Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
