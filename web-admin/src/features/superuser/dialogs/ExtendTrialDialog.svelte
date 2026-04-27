<!-- Confirm-and-submit dialog for extending an org's trial. -->
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
  import { createExtendTrialMutation } from "@rilldata/web-admin/features/superuser/billing/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createEventDispatcher } from "svelte";

  export let open = false;
  export let org: string;
  export let days: number;

  const extendTrial = createExtendTrialMutation();
  const dispatch = createEventDispatcher<{ extended: void }>();

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await $extendTrial.mutateAsync({ data: { org, days } });
      eventBus.emit("notification", {
        type: "success",
        message: `Trial extended by ${days} days for ${org}`,
      });
      dispatch("extended");
      open = false;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to extend trial: ${err}`,
      });
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Extend Trial</AlertDialogTitle>
      <AlertDialogDescription>
        This will extend the trial for "{org}" by {days}
        day{days === 1 ? "" : "s"}.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="primary" onClick={handleConfirm} {loading}
        >Extend Trial</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
