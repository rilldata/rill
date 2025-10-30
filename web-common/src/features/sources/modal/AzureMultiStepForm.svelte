<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { ConnectorDriverPropertyType } from "@rilldata/web-common/runtime-client";
  import { normalizeErrors } from "./utils";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;

  // Derive the connection property from the provided spec (first config property)
  const connectionStringProperty = properties?.[0] ?? {
    key: "connection_string",
    type: ConnectorDriverPropertyType.TYPE_STRING,
    displayName: "Connection string",
    required: false,
    secret: true,
    hint: undefined,
  };
</script>

<!-- Step 1: Connector configuration (Azure) -->
<div>
  <div class="py-1.5 first:pt-0 last:pb-0">
    <Input
      id={connectionStringProperty.key}
      label={connectionStringProperty.displayName ?? "Connection string"}
      placeholder={connectionStringProperty.placeholder}
      optional={!connectionStringProperty.required}
      secret={connectionStringProperty.secret}
      hint={connectionStringProperty.hint}
      errors={normalizeErrors(paramsErrors[connectionStringProperty.key])}
      bind:value={$paramsForm[connectionStringProperty.key]}
      onInput={(_, e) => onStringInputChange(e)}
      alwaysShowError
    />
  </div>
</div>
