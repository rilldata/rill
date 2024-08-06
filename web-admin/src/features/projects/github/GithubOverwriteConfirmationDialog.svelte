<script lang="ts">
  import type { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors";
  import { getRepoNameFromGithubUrl } from "@rilldata/web-common/features/project/github-utils";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";

  export let open = false;
  export let loading: boolean;
  export let error: ReturnType<typeof extractGithubConnectError>;

  export let onConfirm: () => Promise<void>;
  export let onCancel: () => void;
  export let githubUrl: string;
  export let subpath: string;

  let confirmInput = "";
  $: confirmed = confirmInput === "overwrite";

  $: path = `${getRepoNameFromGithubUrl(githubUrl)}/${subpath}`;

  function close() {
    onCancel();
    confirmInput = "";
    open = false;
  }

  async function handleContinue() {
    await onConfirm();
    confirmInput = "";
    open = false;
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle class="flex flex-row gap-x-2 items-center">
        <AlertCircleOutline size="40px" className="text-yellow-600" />
        <div>
          Overwrite files in this {subpath ? "subpath" : "repository"}?
        </div>
      </AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-1.5">
        <div>
          It appears that <b>{path}</b> is not empty. Rill will overwrite this repo’s
          contents with this project.
        </div>
        <div class="mt-1">
          Type <b>overwrite</b> in the box below to confirm:
        </div>
        <Input bind:value={confirmInput} id="confirmation" label="" />
        {#if error?.message}
          <div class="text-red-500 text-sm py-px">
            {error.message}
          </div>
        {/if}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button type="secondary" on:click={close}>Cancel</Button>
      <Button
        type="primary"
        on:click={handleContinue}
        disabled={!confirmed}
        {loading}
      >
        Continue
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
