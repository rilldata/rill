<script lang="ts">
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";

  export let organization: string;
  export let project: string;

  let addUserCommand: string;
  $: addUserCommand = `rill user add --org ${organization} --project ${project} --role viewer`;

  let open = false;
</script>

<Button type="secondary" on:click={() => (open = true)}>Share</Button>

<Dialog
  titleMarginBottomOverride="mb-4"
  widthOverride="w-fit"
  on:close={() => (open = false)}
  {open}
>
  <svelte:fragment slot="title">
    Invite a teammate to your project
  </svelte:fragment>

  <div
    class="flex flex-col gap-y-4 text-left text-sm text-gray-500 w-full"
    slot="body"
  >
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <div>Run this command in the Rill CLI:</div>
        <CLICommandDisplay command={addUserCommand} />
      </svelte:fragment>
      <svelte:fragment slot="read-project">
        <div>
          Ask your organization's admin to invite viewers using the Rill CLI.
        </div>
      </svelte:fragment>
    </ProjectAccessControls>
  </div>

  <svelte:fragment slot="footer">
    <div class="flex mt-4">
      <div class="grow" />
      <Button type="secondary" on:click={() => (open = false)}>Close</Button>
    </div>
  </svelte:fragment>
</Dialog>
