<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import ProjectRenameForm, {
    ProjectRenameFormId,
  } from "@rilldata/web-admin/features/projects/settings/ProjectRenameForm.svelte";
  import { Button } from "@rilldata/web-common/components/button/index.ts";

  let {
    organization,
    project,
    open = $bindable(false),
  }: { organization: string; project: string; open: boolean } = $props();

  let loading = $state(false);
  let changed = $state(false);
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger><div class="hidden"></div></Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Title>Rename Project</Dialog.Title>
    <ProjectRenameForm {organization} {project} bind:loading bind:changed />
    <Dialog.Footer>
      <Dialog.Close>Cancel</Dialog.Close>
      <Button
        submitForm
        form={ProjectRenameFormId}
        type="primary"
        {loading}
        disabled={!changed}
        onRename={() => (open = false)}
      >
        Rename
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
