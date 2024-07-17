<script lang="ts">
  import { createAdminServiceAddProjectMemberUser } from "@rilldata/web-admin/client";
  import { parseError } from "@rilldata/web-admin/features/projects/user-invite/errors";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-invite/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import { slide } from "svelte/transition";

  export let organization: string;
  export let project: string;
  export let onInvite: () => void = () => {};

  const userInvite = createAdminServiceAddProjectMemberUser();

  const initialValues: {
    emails: string[];
    // emails not yet converted to pills
    emailsInput: string;
    role: string;
  } = {
    emails: [],
    emailsInput: "",
    role: "viewer",
  };
  const schema = yup(
    object({
      emails: array(string().email("Invalid email")),
      emailsInput: string(),
      role: string().required(),
    }),
  );

  let submitErrors: string[] = [];

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form, cancel }) {
        if (!form.valid) return;
        const values = form.data;

        if (await inviteUsers(values.emails, values.role)) {
          onInvite();
        } else {
          cancel();
        }
      },
      validationMethod: "oninput",
    },
  );

  function handleSubmit() {
    if ($form.emailsInput !== "") {
      $form.emails = $form.emails.concat(
        ...$form.emailsInput.split(",").map((v) => v.trim()),
      );
      $form.emailsInput = "";
    }
    submit();
  }

  async function inviteUsers(emails: string[], role: string) {
    const succeeded = [];
    const newSubmitErrors: string[] = [];
    await Promise.all(
      emails.map(async (email) => {
        try {
          await $userInvite.mutateAsync({
            organization,
            project,
            data: {
              email,
              role,
            },
          });
          succeeded.push(email);
        } catch (e) {
          newSubmitErrors.push(parseError(e, email));
        }
      }),
    );

    if (succeeded.length) {
      eventBus.emit("notification", {
        type: "success",
        message: `Invited ${succeeded.length} ${succeeded.length === 1 ? "person" : "people"} as ${role}`,
      });
    }
    submitErrors = newSubmitErrors;
    if (newSubmitErrors.length === 0) {
      return true;
    } else {
      $form.emails = $form.emails.filter((e) => !succeeded.includes(e));
      return false;
    }
  }
</script>

<form
  id="user-invite-form"
  on:submit|preventDefault={handleSubmit}
  class="w-full"
  use:enhance
>
  <MultiInput
    id="emails"
    placeholder="Invite by email, separated by commas"
    contentClassName="relative"
    bind:values={$form.emails}
    bind:input={$form.emailsInput}
    errors={$errors.emails}
    useTab
  >
    <div slot="within-input" class="h-full items-center flex">
      <UserRoleSelect bind:value={$form.role} />
    </div>
    <Button
      submitForm
      type="primary"
      form="user-invite-form"
      slot="beside-input"
      loading={$submitting}
      disabled={$form.emailsInput === "" && $form.emails.length === 0}
      forcedStyle="height: 32px !important;"
    >
      Invite
    </Button>
  </MultiInput>
  {#if submitErrors.length}
    <div
      in:slide={{ duration: 200 }}
      class="text-red-500 text-sm py-px flex flex-col"
    >
      {#each submitErrors as error}
        <div>{error}</div>
      {/each}
    </div>
  {/if}
</form>
