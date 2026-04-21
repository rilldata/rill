<!-- Type-to-confirm dialog for redeploying a project. Redeploy is disruptive to
     active users, so it uses the guarded confirmation flow. -->
<script lang="ts">
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { createRedeployProjectMutation } from "@rilldata/web-admin/features/superuser/projects/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let open = false;
  export let projectPath: string; // "org/project"

  const redeploy = createRedeployProjectMutation();

  let loading = false;
  let error: string | undefined = undefined;

  async function handleConfirm() {
    const [org, project] = projectPath.split("/");
    loading = true;
    error = undefined;
    try {
      await $redeploy.mutateAsync({ org, project });
      eventBus.emit("notification", {
        type: "success",
        message: `Project ${projectPath} redeployed`,
      });
    } catch (err) {
      error = `Failed: ${err}`;
      throw err;
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialogGuardedConfirmation
  bind:open
  title="Redeploy Project"
  description={`This will completely redeploy ${projectPath}. This is disruptive to active users.`}
  confirmText={projectPath}
  confirmButtonText="Redeploy"
  confirmButtonType="destructive"
  {loading}
  {error}
  onConfirm={handleConfirm}
/>
