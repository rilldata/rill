<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import {
    Button,
    type ButtonType,
  } from "@rilldata/web-common/components/button/index.js";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";

  export let open = false;

  export let title: string;
  export let description: string;
  export let confirmText: string;
  export let confirmButtonText: string = "Continue";
  export let confirmButtonType: ButtonType = "primary";

  export let loading: boolean;
  export let error: string | undefined = undefined;
  export let onConfirm: () => Promise<void>;
  export let onCancel: () => void = () => {};

  let confirmInput = "";
  $: confirmed = confirmInput === confirmText;
  $: iconColor =
    confirmButtonType === "destructive" ? "text-red-500" : "text-yellow-600";

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
  <AlertDialogTrigger asChild let:builder>
    <slot {builder} />
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle class="flex flex-row gap-x-2 items-center">
        <AlertCircleOutline size="40px" className={iconColor} />
        <div>{title}</div>
      </AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-1.5">
        <div>{description}</div>
        <div class="mt-1">
          Type <b>{confirmText}</b> in the box below to confirm:
        </div>
        <Input bind:value={confirmInput} id="confirmation" label="" />
        {#if error}
          <div class="text-red-500 text-sm py-px">
            {error}
          </div>
        {/if}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button
        type={confirmButtonType === "destructive" ? "tertiary" : "secondary"}
        onClick={close}
      >
        Cancel
      </Button>
      <Button
        type={confirmButtonType}
        onClick={handleContinue}
        disabled={!confirmed}
        {loading}
      >
        {confirmButtonText}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
