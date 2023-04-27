<script lang="ts">
  import AccessControls from "@rilldata/web-admin/components/access-controls/AccessControls.svelte";

  export let organization: string;
  export let project: string;

  let addUserCommand: string;
  $: addUserCommand = `rill user add --org ${organization} --project ${project} --role viewer`;
</script>

<AccessControls {organization} {project}>
  <svelte:fragment slot="manage-project">
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
      >Share</span
    >
    <div>
      Run this command in the Rill CLI to invite a teammate to view this
      project.
    </div>
    <div class="border bg-gray-50 p-0.5 h-6 text-gray-400">
      <section class="">
        {addUserCommand}
      </section>
    </div>
  </svelte:fragment>
  <svelte:fragment slot="read-project">
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
      >Share</span
    >
    <div>
      Ask your organization’s admin to invite viewers using the Rill CLI.
    </div>
  </svelte:fragment>
</AccessControls>
