<!-- Confirm-to-hibernate dialog. Hibernate is reversible (the data is preserved
     and the project can be redeployed later), so a single-click confirm is
     appropriate. -->
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
  import { createHibernateProjectMutation } from "@rilldata/web-admin/features/superuser/projects/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let open = false;
  export let projectPath: string; // "org/project"

  const hibernate = createHibernateProjectMutation();

  let loading = false;

  async function handleConfirm() {
    const [org, project] = projectPath.split("/");
    loading = true;
    try {
      await $hibernate.mutateAsync({ org, project });
      eventBus.emit("notification", {
        type: "success",
        message: `Project ${projectPath} hibernated`,
      });
      open = false;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed: ${err}`,
      });
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Hibernate Project</AlertDialogTitle>
      <AlertDialogDescription>
        This will hibernate the deployment for {projectPath}. The project data
        will be preserved but the deployment will be stopped.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="primary" onClick={handleConfirm} {loading}>Hibernate</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
