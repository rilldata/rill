<script context="module" lang="ts">
  export const CreateNewOrgFormId = "org-name-form";
</script>

<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { sanitizeSlug } from "@rilldata/web-common/lib/string-utils";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  // We need different sizes for showing in dialog vs a full page form.
  // "lg" matches all the input sizes so we have "lg"/"xl" and not something else.
  export let size: "lg" | "xl" = "lg";
  export let createOrg: (name: string, displayName: string) => Promise<void>;

  const initialValues: {
    name: string;
    displayName: string;
  } = {
    name: "",
    displayName: "",
  };
  const schema = yup(
    object({
      name: string()
        .required()
        // from admin/databases/database.go::InsertOrganizationOptions::Name
        .matches(
          /^[_a-zA-Z0-9][-_a-zA-Z0-9]*$/,
          "name must only have alphabets and numbers",
        )
        .min(2, "name must be at least 2 characters long")
        .max(40, "name must be at most 40 characters long"),
      displayName: string(),
    }),
  );

  const { form, errors, enhance, submi } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      onUpdate: async ({ form }) => {
        if (!form.valid) return;
        const values = form.data;

        await createOrg(values.name, values.displayName);
      },
      onError({ result }) {
        // Mapping for backend error to a more user friendly UI error message.
        if (
          result.error.message.includes("an org with that name already exists")
        ) {
          $errors["name"] = [
            `Organization name '${$form.name}' is already taken. Please try a different name.`,
          ];
        }
      },
    },
  );

  // As a convenience, we auto generate an org name based on the display name.
  // But the moment the org name is directly changed,
  // we should stop doing this since the user probably changed it directly.
  let orgNameChangedDirectly = false;
  function updateName(displayName: string) {
    if (orgNameChangedDirectly) return;
    $form.name = sanitizeSlug(displayName);
  }
  $: updateName($form.displayName!);
</script>

<form
  id={CreateNewOrgFormId}
  onsubmit={(e) => {
    e.preventDefault();
    submit(e);
  }}
  use:enhance
  class="flex flex-col gap-y-3"
>
  <Input
    bind:value={$form.displayName}
    errors={$errors?.displayName}
    id="displayName"
    label="Organization display name"
    textClass="text-sm"
    width="500px"
    {size}
    capitalizeLabel={false}
  />
  <Input
    bind:value={$form.name}
    errors={$errors?.name}
    id="name"
    label="URL"
    textClass="text-sm"
    alwaysShowError
    width="500px"
    {size}
    onInput={() => (orgNameChangedDirectly = true)}
    textInputPrefix="https://ui.rilldata.com/"
  >
    <!-- TODO: once we have the path to docs we can add this back -->
    <!--    <div class="text-xs text-left" slot="description">-->
    <!--      Must comply with <a-->
    <!--        href="https://docs.rilldata.com/reference/cli/org/create#TODO"-->
    <!--        target="_blank">our naming rules.</a-->
    <!--      >-->
    <!--    </div>-->
  </Input>
</form>
