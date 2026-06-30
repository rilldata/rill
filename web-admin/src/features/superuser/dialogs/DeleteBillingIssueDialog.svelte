<!-- Confirm dialog for deleting a single billing issue from an org. -->
<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    getAdminServiceListOrganizationBillingIssuesQueryKey,
    type V1BillingIssueType,
  } from "@rilldata/web-admin/client";
  import { createDeleteBillingIssueMutation } from "@rilldata/web-admin/features/superuser/billing/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let org: string;
  export let issueType: V1BillingIssueType;

  const deleteIssue = createDeleteBillingIssueMutation();
  const queryClient = useQueryClient();

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await $deleteIssue.mutateAsync({ org, type: issueType });
      eventBus.emit("notification", {
        type: "success",
        message: `Billing issue "${issueType}" deleted for ${org}`,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationBillingIssuesQueryKey(org),
      });
      open = false;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to delete billing issue: ${err}`,
      });
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete Billing Issue</AlertDialogTitle>
      <AlertDialogDescription>
        This will delete the billing issue "{issueType}" for "{org}". This
        action cannot be undone.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="destructive" onClick={handleConfirm} {loading}
        >Delete</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
