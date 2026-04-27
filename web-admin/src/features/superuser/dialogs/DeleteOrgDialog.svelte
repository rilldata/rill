<!-- Type-to-confirm dialog for deleting an organization. -->
<script lang="ts">
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { createDeleteOrgMutation } from "@rilldata/web-admin/features/superuser/organizations/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createEventDispatcher } from "svelte";

  export let open = false;
  export let org: string;

  const deleteOrg = createDeleteOrgMutation();
  const dispatch = createEventDispatcher<{ deleted: { org: string } }>();

  let loading = false;
  let error: string | undefined = undefined;

  async function handleConfirm() {
    loading = true;
    error = undefined;
    try {
      await $deleteOrg.mutateAsync({ org });
      eventBus.emit("notification", {
        type: "success",
        message: `Organization "${org}" deleted`,
      });
      dispatch("deleted", { org });
    } catch (err) {
      error = `Failed to delete organization: ${err}`;
      throw err;
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialogGuardedConfirmation
  bind:open
  title="Delete Organization"
  description={`This will permanently delete "${org}" and all its projects, members, and data. This action cannot be undone.`}
  confirmText={org}
  confirmButtonText="Delete"
  confirmButtonType="destructive"
  {loading}
  {error}
  onConfirm={handleConfirm}
/>
