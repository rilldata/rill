<script lang="ts">
  import { createForm } from "svelte-forms-lib";
  import type { Writable } from "svelte/store";
  import * as yup from "yup";
  import {
    ConnectorSpec,
    GCS,
    GCSYupSchema,
    HTTP,
    HTTPYupSchema,
    S3,
    S3YupSchema,
  } from "../../../connectors/schemas";
  import { Button } from "../../button";
  import Input from "../../Input.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";

  export let connector;
  export let connectorDescription = "";

  let connectorSpec: ConnectorSpec;
  let yupSchema: yup.AnyObjectSchema;

  // state from svelte-forms-lib
  let form: Writable<any>;
  let errors: Writable<Record<never, string>>;
  let handleSubmit: (event: Event) => any;

  function extendYupSchemaWithSourceName(yupSchema: yup.AnyObjectSchema) {
    return yupSchema.concat(
      yup.object().shape({
        sourceName: yup.string().required(),
      })
    );
  }

  function compileCreateSourceSql(values) {
    const compiledKeyValues = Object.entries(values)
      .map(([key, value]) => `${key}='${value}'`)
      .join(", ");

    return (
      `CREATE SOURCE ${values.sourceName} WITH (connector = '${connectorSpec.name}', ` +
      compiledKeyValues +
      `)`
    );
  }

  function onConnectorChange(connector: string) {
    switch (connector) {
      case "S3":
        connectorSpec = S3;
        yupSchema = S3YupSchema;
        break;
      case "GCS":
        connectorSpec = GCS;
        yupSchema = GCSYupSchema;
        break;
      case "HTTP":
        connectorSpec = HTTP;
        yupSchema = HTTPYupSchema;
        break;
      default:
        throw new Error("Unknown connector");
    }

    yupSchema = extendYupSchemaWithSourceName(yupSchema);
    connectorDescription = connectorSpec.description;

    ({ form, errors, handleSubmit } = createForm({
      // TODO: initialValues should come from SQL asset and be reactive to asset modifications
      initialValues: {},
      validationSchema: yupSchema,
      onSubmit: (values) => {
        const sql = compileCreateSourceSql(values);
        // TODO: dispatch sql to SQL editor
        // TODO: submit sql to backend
        alert(sql);
      },
    }));
  }

  $: onConnectorChange(connector);
</script>

<div class="px-4 flex-grow overflow-y-auto pb-2">
  <form on:submit={handleSubmit} id="remote-source-{connector}-form">
    <div class="py-2">
      <Input
        label="Source name"
        bind:value={$form["sourceName"]}
        error={$errors["sourceName"]}
        placeholder="my_new_source"
      />
    </div>
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
        {#if attributes.type === "checkbox"}
          <label for={name} class="flex items-center">
            <input
              id={name}
              type="checkbox"
              bind:checked={$form[name]}
              class="h-5 w-5"
            />
            <span class="ml-2 text-sm">{label}</span>
          </label>
        {/if}
      </div>
    {/each}
  </form>
</div>
<div class="bg-gray-100 border-t border-gray-300">
  <DialogFooter>
    <Button type="primary" submitForm form="remote-source-{connector}-form">
      Add source
    </Button>
  </DialogFooter>
</div>
