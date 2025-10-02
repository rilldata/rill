<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";
  import { inferSourceName } from "../sourceUtils";

  export let properties: any[];
  export let formId: string;
  export let form: any;
  export let errors: any;
  export let enhance: any;
  export let submit: any;
  export let onStringInputChange: (event: Event) => void;

  // Filter properties based on connector type
  $: filteredProperties = (() => {
    // FIXME: https://linear.app/rilldata/issue/APP-408/support-ducklake-in-the-ui
    if (properties.some((p) => p.key === "attach" || p.key === "mode")) {
      return properties.filter(
        (property) => property.key !== "attach" && property.key !== "mode",
      );
    }
    // For other connectors, filter out noPrompt properties
    return properties.filter((property) => !property.noPrompt);
  })();
</script>

<form
  id={formId}
  class="pb-5 flex-grow overflow-y-auto"
  use:enhance
  on:submit|preventDefault={submit}
>
  {#each filteredProperties as property (property.key)}
    {@const propertyKey = property.key ?? ""}
    {@const label =
      property.displayName + (property.required ? "" : " (optional)")}
    <div class="py-1.5 first:pt-0 last:pb-0">
      {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
        <Input
          id={propertyKey}
          label={property.displayName}
          placeholder={property.placeholder}
          optional={!property.required}
          secret={property.secret}
          hint={property.hint}
          errors={normalizeErrors(errors[propertyKey])}
          bind:value={$form[propertyKey]}
          onInput={(_, e) => onStringInputChange(e)}
          alwaysShowError
        />
      {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
        <Checkbox
          id={propertyKey}
          bind:checked={$form[propertyKey]}
          {label}
          hint={property.hint}
          optional={!property.required}
        />
      {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
        <InformationalField
          description={property.description}
          hint={property.hint}
          href={property.docsUrl}
        />
      {/if}
    </div>
  {/each}
</form>
