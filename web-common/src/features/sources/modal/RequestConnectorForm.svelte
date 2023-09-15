<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";

  const dispatch = createEventDispatcher();

  const { form, errors, handleChange, handleSubmit, isSubmitting } = createForm(
    {
      initialValues: {},
      onSubmit: async (values) => {
        // TODO: create a Google Form for this
        console.log("submitting form", values);
        dispatch("close");
      },
    }
  );
</script>

<div class="flex flex-col">
  <form on:submit|preventDefault={handleSubmit} id="request-connector-form">
    <span>
      Don't see the connector you're looking for? Let us know what we're
      missing!
    </span>

    <div class="pt-2 pb-4">
      <Input
        id="request"
        label="Connector"
        placeholder="Your data source"
        error={$errors["request"]}
        bind:value={$form["request"]}
        on:change={handleChange}
      />
    </div>
  </form>
  <div class="flex">
    <div class="grow" />
    <Button
      type="primary"
      submitForm
      form="request-connector-form"
      disabled={$isSubmitting}
    >
      Request connector
    </Button>
  </div>
</div>
