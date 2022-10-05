<script lang="ts">
  import { createForm } from "svelte-forms-lib";
  import type * as yup from "yup";
  import type { ConnectorSpec } from "../../../connectors/schemas";
  import Button from "../../button/Button.svelte";
  import Input from "../../Input.svelte";

  export let sourceName: string;
  export let connectorSpec: ConnectorSpec;
  export let yupSchema: yup.AnyObjectSchema;

  function compileCreateSourceSql(values) {
    const compiledKeyValues = Object.entries(values)
      .map(([key, value]) => `${key}='${value}'`)
      .join(", ");

    return (
      `CREATE SOURCE ${sourceName} WITH (connector = '${connectorSpec.name}', ` +
      compiledKeyValues +
      `)`
    );
  }

  const { form, errors, handleSubmit } = createForm({
    // TODO: initialValues should come from SQL asset and be reactive to asset modifications
    initialValues: {},
    validationSchema: yupSchema,
    onSubmit: (values) => {
      const sql = compileCreateSourceSql(values);
      // TODO: dispatch sql to SQL editor
      // TODO: submit sql to backend
      alert(sql);
    },
  });
</script>

<div class="max-w-sm">
  <h1>{connectorSpec.title}</h1>
  <div>{@html connectorSpec.description}</div>
  <form on:submit={handleSubmit}>
    <div class="py-4">
      {#each Object.entries(connectorSpec.fields) as [name, attributes]}
        {@const label =
          attributes.label + (attributes.required ? "" : " (optional)")}
        <div class="py-2">
          {#if attributes.type === "text"}
            <Input
              id={name}
              {label}
              error={$errors[name]}
              bind:value={$form[name]}
              placeholder={attributes.placeholder}
            />
          {/if}
        </div>
      {/each}
    </div>

    <Button type="primary" submitForm>Submit</Button>
  </form>
</div>
