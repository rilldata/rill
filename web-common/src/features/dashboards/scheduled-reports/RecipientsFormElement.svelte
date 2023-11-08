<script lang="ts">
  import { useProjectMembersEmails } from "@rilldata/web-admin/features/projects/selectors";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import MultiselectCombobox from "../../../components/forms/MultiSelectCombobox.svelte";

  export let organization: string;
  export let project: string;
  export let recipients: string[];
  export let error: string;

  $: membersEmails = useProjectMembersEmails(organization, project);

  function removeRecipient(recipient: string) {
    recipients = recipients.filter((r) => r !== recipient);
  }
</script>

<div>
  <MultiselectCombobox
    bind:selectedValues={recipients}
    id="recipients"
    label="Recipients"
    {error}
    placeholder="Search emails"
    options={$membersEmails.data}
    hint="Recipients may receive different views based on their security policy."
  />
  <!-- Project members to invite -->
  <ul class="flex flex-col gap-y-2 my-5 max-h-[130px] overflow-y-auto">
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
