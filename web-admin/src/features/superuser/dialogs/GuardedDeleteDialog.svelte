<!-- Type-to-confirm destructive dialog for superuser actions -->
<script lang="ts">
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";

  export let open = false;
  export let title: string;
  export let description: string;
  export let confirmText: string;
  export let confirmButtonText: string = "Delete";
  export let loading = false;
  export let error: string | undefined = undefined;
  export let onConfirm: () => Promise<void>;
</script>

<AlertDialogGuardedConfirmation
  bind:open
  {title}
  {description}
  {confirmText}
  {confirmButtonText}
  confirmButtonType="destructive"
  {loading}
  {error}
  {onConfirm}
>
  <slot />
</AlertDialogGuardedConfirmation>
