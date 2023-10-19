<script lang="ts">
  import { createAdminServiceListProjectMembers } from "@rilldata/web-admin/client";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import FormItemMultiselectCombobox from "../../../components/forms/FormItemMultiSelectCombobox.svelte";

  export let organization: string;
  export let project: string;
  export let recipients: string[];

  $: projectMembersQuery = createAdminServiceListProjectMembers(
    organization,
    project
  );
  $: projectMembers = $projectMembersQuery.data?.members ?? [];

  function removeRecipient(recipient: string) {
    recipients = recipients.filter((r) => r !== recipient);
  }
</script>

<div>
  <FormItemMultiselectCombobox
    bind:selectedValues={recipients}
    id="recipients"
    label="Recipients"
    placeholder="Search emails"
    options={projectMembers.map((member) => member.userEmail)}
  />
  <span class="text-gray-500 text-sm py-px leading-snug">
    Recipients may receive different views based on their security policy.
  </span>
  <!-- Project members to invite -->
  <ul class="py-5 flex flex-col gap-y-2">
    {#if recipients.length > 0}
      {#each recipients as recipient}
        <div class="flex items-center justify-between group">
          <div class="flex gap-x-2 items-center">
            <div
              class="w-8 h-8 rounded-full bg-red-200 grid place-items-center"
            >
              <span class="text-orange-600">{recipient[0].toUpperCase()}</span>
            </div>
            <li class="text-gray-700 text-sm">{recipient}</li>
          </div>
          <div
            on:click={() => removeRecipient(recipient)}
            on:keydown={() => removeRecipient(recipient)}
            class="invisible group-hover:visible cursor-pointer"
          >
            <Close size="24px" className="text-gray-500" />
          </div>
        </div>
      {/each}
    {/if}
  </ul>
</div>
