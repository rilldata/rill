<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let open = false;
  export let groupName: string;
  export let onRename: (groupName: string, newName: string) => void;

  const formId = "rename-user-group-form";

  const initialValues = {
    newName: groupName,
  };

  const schema = yup(
    object({
      newName: string()
        .required("New user group name is required")
        .min(3, "New user group name must be at least 3 characters")
        .matches(
          /^[a-z0-9]+(-[a-z0-9]+)*$/,
          "New user group name must be lowercase and can contain letters, numbers, and hyphens (slug)",
        ),
    }),
  );

  const { form, enhance, submit, errors, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        try {
          await onRename(groupName, values.newName);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Rename user group</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      class="w-full"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <div class="flex flex-col gap-2 w-full">
        <Input
          bind:value={$form.newName}
          placeholder="New user group name"
          errors={$errors.newName}
          alwaysShowError
        />
      </div>
    </form>
    <DialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button
        type="primary"
        disabled={$submitting || $form.newName.trim() === groupName}
        form={formId}
        submitForm
      >
        Rename
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
