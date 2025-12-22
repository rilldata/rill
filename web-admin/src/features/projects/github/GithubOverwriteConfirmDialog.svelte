<script lang="ts">
  import { getRepoNameFromGitRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
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
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";

  export let open = false;
  export let loading: boolean;
  export let error: string | undefined;

  export let onConfirm: () => Promise<void>;
  export let onCancel: () => void = () => {};
  export let githubRemote: string;
  export let subpath: string;

  $: path =
    `${getRepoNameFromGitRemote(githubRemote)}` +
    (subpath ? `/${subpath}` : "");

  const CONFIRMATION_TEXT = "overwrite";

  let confirmInput = "";
  $: confirmed = confirmInput === CONFIRMATION_TEXT;

  function close() {
    onCancel();
    open = false;
  }

  async function handleContinue() {
    await onConfirm();
  }
</script>

<AlertDialog
  bind:open
  onOpenChange={(o) => {
    if (o) {
      confirmInput = "";
    }
  }}
>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle class="flex flex-row gap-x-2 items-center">
        <AlertCircleOutline size="40px" className="text-yellow-600" />
        <div>Pull changes from {path}?</div>
      </AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-1.5">
        <div>
          Current project contents will be overwritten with the contents of the
          repository. There is no way to retrieve this current project. Are you
          sure?
        </div>
        <div class="mt-1">
          Type <b>{CONFIRMATION_TEXT}</b> in the box below to confirm:
        </div>
        <Input bind:value={confirmInput} id="confirmation" label="" />
        {#if error}
          <div class="text-red-500 text-sm py-px">{error}</div>
        {/if}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button type="secondary" onClick={close}>Cancel</Button>
      <Button
        type="primary"
        onClick={handleContinue}
        disabled={!confirmed}
        {loading}
      >
        Continue
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
