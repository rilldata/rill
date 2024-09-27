<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  export let open = false;
  export let groupName: string;
  export let onCreate: (newName: string) => void;

  const formId = "add-user-group-form";

  const initialValues = {
    newName: "",
  };

  const schema = yup(
    object({
      newName: string()
        .required("User group name is required")
        .min(3, "User group name must be at least 3 characters")
        .matches(
          /^[a-z0-9]+(-[a-z0-9]+)*$/,
          "User group name must be lowercase and can contain letters, numbers, and hyphens (slug)",
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
          await onCreate(values.newName);
          open = false;
        } catch (error) {
          console.error(error);
        }
      },
    },
  );

  $: console.log($errors);
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
    groupName = "";
  }}
  onOpenChange={(open) => {
    if (!open) {
      groupName = "";
    }
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add user group</DialogTitle>
    </DialogHeader>
    <DialogFooter class="mt-4">
      <form
        id={formId}
        class="w-full"
        on:submit|preventDefault={submit}
        use:enhance
      >
        <div class="flex flex-col gap-2 w-full">
          <Input
            bind:value={$form.newName}
            placeholder="User group name"
            errors={$errors.newName}
            alwaysShowError
          />
          <Button
            type="primary"
            large
            disabled={$submitting || $form.newName.trim() === ""}
            form={formId}
            submitForm
          >
            Create
          </Button>
        </div>
      </form>
    </DialogFooter>
  </DialogContent>
</Dialog>
