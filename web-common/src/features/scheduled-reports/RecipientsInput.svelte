<script lang="ts">
  import InputArray from "@rilldata/web-common/components/forms/InputArray.svelte";

  export let formState: any; // svelte-forms-lib's FormState
  export let hint: string;

  const { form, errors } = formState;

  // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
  // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
  $: recipientErrors = $errors.recipients as unknown as { email: string }[];
</script>

<InputArray
  accessorKey="email"
  addItemLabel="Add email"
  bind:errors={recipientErrors}
  bind:values={$form["recipients"]}
  {hint}
  id="recipients"
  label="Recipients"
  on:add-item={() => {
    $form["recipients"] = $form["recipients"].concat({ email: "" });
    recipientErrors = recipientErrors.concat({ email: "" });

    // Focus on the new input element
    setTimeout(() => {
      const input = document.getElementById(
        `recipients.${$form["recipients"].length - 1}.email`,
      );
      input?.focus();
    }, 0);
  }}
  on:remove-item={(event) => {
    const index = event.detail.index;
    $form["recipients"] = $form["recipients"].filter((r, i) => i !== index);
    recipientErrors = recipientErrors.filter((r, i) => i !== index);
  }}
  placeholder="Enter an email address"
/>
