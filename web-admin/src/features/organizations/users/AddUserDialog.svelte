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
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string, boolean } from "yup";
  import { capitalize } from "@rilldata/web-common/components/table/utils";

  export let open = false;
  export let email: string;
  export let role: string;
  export let isSuperUser: boolean;
  export let onCreate: (
    newEmail: string,
    newRole: string,
    isSuperUser: boolean,
  ) => void;

  const roleOptions = ["admin", "viewer", "collaborator"].map((value) => ({
    value,
    label: capitalize(value),
  }));

  const formId = "add-user-form";

  const initialValues = {
    newEmail: "",
    newRole: "",
    isSuperUser: false,
  };

  const schema = yup(
    object({
      newEmail: string()
        .email("Invalid email address")
        .required("Email is required"),
      newRole: string().required("Role is required"),
      isSuperUser: boolean().optional(),
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

        const { newEmail, newRole, isSuperUser } = values;

        try {
          await onCreate(newEmail, newRole, isSuperUser);
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
    email = "";
    role = "";
    isSuperUser = false;
  }}
  onOpenChange={(open) => {
    if (!open) {
      email = "";
      role = "";
      isSuperUser = false;
    }
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add user</DialogTitle>
    </DialogHeader>
    <DialogFooter class="mt-2">
      <form
        id={formId}
        class="w-full"
        on:submit|preventDefault={submit}
        use:enhance
      >
        <div class="flex flex-col gap-3 w-full">
          <Input
            bind:value={$form.newEmail}
            id="newEmail"
            label="Email"
            placeholder="Email"
            errors={$errors.newEmail}
            alwaysShowError
          />
          <Select
            bind:value={$form.newRole}
            id="newRole"
            label="Role"
            placeholder="Role"
            options={roleOptions}
          />
          <!-- TODO: add checkbox for superuser -->
        </div>
        <Button
          type="primary"
          large
          disabled={$submitting ||
            $form.newEmail.trim() === "" ||
            $form.newRole === ""}
          form={formId}
          submitForm
          class="w-full mt-4"
        >
          Create
        </Button>
      </form>
    </DialogFooter>
  </DialogContent>
</Dialog>
