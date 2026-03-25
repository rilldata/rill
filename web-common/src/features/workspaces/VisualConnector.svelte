<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    getConnectorSchema,
    getSchemaNameFromDriver,
    toConnectorDriver,
    multiStepFormSchemas,
  } from "@rilldata/web-common/features/sources/modal/connector-schemas";
  import {
    compileConnectorYAML,
    replaceOrAddEnvVariable,
    makeEnvVarKey,
  } from "@rilldata/web-common/features/connectors/code-utils";
  import { loadConnectorFormValues } from "@rilldata/web-common/features/sources/modal/edit-connector-utils";
  import JSONSchemaFormRenderer from "@rilldata/web-common/features/templates/JSONSchemaFormRenderer.svelte";
  import {
    getSchemaSecretKeys,
    getSchemaStringKeys,
    getSchemaFieldMetaList,
    getSchemaInitialValues,
    filterSchemaValuesForSubmit,
    getConditionalValues,
  } from "@rilldata/web-common/features/templates/schema-utils";
  import { processFileContent } from "@rilldata/web-common/features/templates/file-encoding";
  import { ICONS } from "@rilldata/web-common/features/sources/modal/icons";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    runtimeServiceGetFile,
    runtimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { parse } from "yaml";
  import { writable, get } from "svelte/store";
  import { tick } from "svelte";
  import { XIcon } from "lucide-svelte";

  export let fileArtifact: FileArtifact;

  const runtimeClient = useRuntimeClient();

  $: ({ remoteContent } = fileArtifact);

  // Parse driver from YAML to determine which schema to use
  let driver = "";
  $: {
    try {
      const obj = parse($remoteContent ?? "");
      driver = obj?.driver ?? "";
    } catch {
      driver = "";
    }
  }

  // Connector instance name from the file path
  $: connectorName = fileArtifact.path
    .replace(/^\/connectors\//, "")
    .replace(/\.yaml$/, "");

  // Prefer connector name as schema key (handles motherduck→duckdb, supabase→postgres
  // where the file name matches the schema but the driver field is the parent driver).
  // Fall back to driver-based lookup for generic connectors (e.g. "my-pg" → postgres).
  $: schemaName =
    connectorName in multiStepFormSchemas
      ? connectorName
      : getSchemaNameFromDriver(driver);
  $: baseSchema = getConnectorSchema(schemaName);
  $: connectorDriver = toConnectorDriver(schemaName);
  $: secretKeysList = baseSchema
    ? getSchemaSecretKeys(baseSchema, { step: "connector" })
    : [];
  $: secretKeys = new Set(secretKeysList);

  // Enrich schema with actual env var names from YAML template refs
  // (e.g. show GOOGLE_APPLICATION_CREDENTIALS_2 instead of the base name)
  $: schema = enrichSchemaEnvVarNames(baseSchema, $remoteContent);
  $: stringKeysList = schema
    ? getSchemaStringKeys(schema, { step: "connector" })
    : [];
  $: schemaFields = schema
    ? getSchemaFieldMetaList(schema, { step: "connector" })
    : [];

  // Form store compatible with JSONSchemaFormRenderer's SuperFormStore interface
  const formStore = writable<Record<string, unknown>>({});
  const errorsStore = writable<Record<string, string[]>>({});

  // Track whether we're loading to suppress dirty tracking
  let suppressSync = false;
  // Track whether the form has been modified since load
  let formDirty = false;

  const RESERVED_KEYS = new Set(["type", "driver", "managed"]);

  // Load form values from resource + .env (resolves secret template refs)
  let formLoaded = false;
  $: if (schema && connectorName && schemaName) {
    loadFormFromResource(connectorName, schemaName);
  }

  async function loadFormFromResource(name: string, sName: string) {
    suppressSync = true;

    // Read .env to resolve secret template refs
    let envBlob = "";
    try {
      const envFile = await runtimeServiceGetFile(runtimeClient, {
        path: ".env",
      });
      envBlob = envFile.blob || "";
    } catch {
      // .env may not exist
    }

    let loaded: Record<string, unknown> = {};
    try {
      loaded = await loadConnectorFormValues(
        runtimeClient,
        name,
        sName,
        envBlob,
      );
    } catch {
      // Resource may be in error state (e.g. non-standard keys);
      // fall back to parsing YAML directly for non-secret fields
      // and resolving secret template refs from .env
      loaded = parseYamlWithEnv($remoteContent ?? "", envBlob);
    }

    // Merge schema defaults (e.g. connector_type: "rill-managed") as base,
    // then overlay loaded values so existing config takes precedence
    const defaults = schema
      ? getSchemaInitialValues(schema, { step: "connector" })
      : {};
    formStore.set({ ...defaults, ...loaded });
    formLoaded = true;
    formDirty = false;
    await tick();
    suppressSync = false;
  }

  // Parse YAML + resolve secret template refs from .env as a fallback
  function parseYamlWithEnv(
    content: string,
    envBlob: string,
  ): Record<string, unknown> {
    const envMap: Record<string, string> = {};
    for (const line of envBlob.split("\n")) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;
      const eq = trimmed.indexOf("=");
      if (eq === -1) continue;
      envMap[trimmed.slice(0, eq)] = trimmed.slice(eq + 1);
    }

    try {
      const obj = parse(content) ?? {};
      const values: Record<string, unknown> = {};
      for (const [key, value] of Object.entries(obj)) {
        if (RESERVED_KEYS.has(key)) continue;
        if (
          secretKeys.has(key) &&
          typeof value === "string" &&
          value.includes("{{ .env.")
        ) {
          // Resolve template ref from .env
          const envVar = extractEnvVar(value);
          if (envVar && envMap[envVar] !== undefined) {
            values[key] = envMap[envVar];
          }
        } else {
          values[key] = value;
        }
      }
      return values;
    } catch {
      return {};
    }
  }

  // Superform-compatible store wrapper for JSONSchemaFormRenderer
  // YAML is only updated on save, not on every form change.
  const formStoreWrapper = {
    subscribe: formStore.subscribe,
    set: (value: Record<string, unknown>) => {
      formStore.set(value);
      if (!suppressSync) formDirty = true;
    },
    update: (
      updater: (value: Record<string, unknown>) => Record<string, unknown>,
      _options?: { taint?: boolean },
    ) => {
      formStore.update(updater);
      if (!suppressSync) formDirty = true;
    },
  };

  function handleStringInputChange(_e: Event) {
    // no-op: YAML updates on save only
  }

  async function handleFileUpload(
    file: File,
    fieldKey: string,
  ): Promise<string> {
    const content = await file.text();

    if (fieldKey && schema) {
      const field = schema.properties?.[fieldKey];
      if (field?.["x-file-encoding"]) {
        const result = processFileContent(content, field);

        // Extract values from file (e.g. project_id from GCP JSON)
        if (Object.keys(result.extractedValues).length > 0) {
          formStoreWrapper.update(($form) => {
            for (const [key, value] of Object.entries(
              result.extractedValues,
            )) {
              $form[key] = value;
            }
            return $form;
          });
        }

        return result.encodedContent;
      }
    }

    return content;
  }

  // Create a schema copy with x-env-var-name overridden to reflect the
  // actual env var names used in the YAML (e.g. GOOGLE_APPLICATION_CREDENTIALS_2)
  function enrichSchemaEnvVarNames(
    s: typeof baseSchema,
    yamlContent: string | null | undefined,
  ): typeof baseSchema {
    if (!s?.properties || !yamlContent) return s;
    try {
      const obj = parse(yamlContent) ?? {};
      let modified = false;
      const newProps: Record<string, any> = {};
      for (const [key, prop] of Object.entries(s.properties)) {
        const envVar = extractEnvVar(obj[key]);
        if (envVar && envVar !== prop["x-env-var-name"]) {
          newProps[key] = { ...prop, "x-env-var-name": envVar };
          modified = true;
        } else {
          newProps[key] = prop;
        }
      }
      return modified ? { ...s, properties: newProps } : s;
    } catch {
      return s;
    }
  }

  let saving = false;

  // Extract env var name from a Go template ref like {{ .env.POSTGRES_DSN }}
  function extractEnvVar(value: unknown): string | undefined {
    if (typeof value !== "string") return undefined;
    const match = value.match(/\{\{\s*\.env\.(\w+)\s*\}\}/);
    return match?.[1];
  }

  async function save() {
    if (!connectorDriver || !schema) {
      fileArtifact.saveLocalContent(true);
      return;
    }

    saving = true;
    try {
      // Filter out x-ui-only fields and inactive tab group fields
      // (e.g. when using DSN, strip host/port/sslmode and vice versa)
      const rawValues = get(formStore);
      const values = schema
        ? {
            ...filterSchemaValuesForSubmit(schema, rawValues, {
              step: "connector",
            }),
            // Re-apply enforced const values from conditionals (e.g. managed: true)
            // which may have been stripped as x-ui-only
            ...getConditionalValues(schema, rawValues),
          }
        : rawValues;

      // Read existing .env
      let envBlob = "";
      try {
        const envFile = await runtimeServiceGetFile(runtimeClient, {
          path: ".env",
        });
        envBlob = envFile.blob || "";
      } catch {
        // .env may not exist
      }

      // Parse existing YAML to find current env var names for each secret
      const existingYaml = parse($remoteContent ?? "") ?? {};
      const existingEnvVars: Record<string, string> = {};
      for (const key of secretKeysList) {
        const envVar = extractEnvVar(existingYaml[key]);
        if (envVar) existingEnvVars[key] = envVar;
      }

      // Update .env: reuse existing env var names, only generate new ones for new secrets
      for (const key of secretKeysList) {
        const val = values[key];
        if (!val || (typeof val === "string" && !val.trim())) continue;

        const envVarName =
          existingEnvVars[key] ??
          makeEnvVarKey(
            connectorDriver.name as string,
            key,
            envBlob,
            schema ?? undefined,
          );
        existingEnvVars[key] = envVarName;
        envBlob = replaceOrAddEnvVariable(envBlob, envVarName, val as string);
      }

      // Write .env file
      await runtimeServicePutFile(runtimeClient, {
        path: ".env",
        blob: envBlob,
        create: true,
        createOnly: false,
      });

      // Build YAML without existingEnvBlob so compileConnectorYAML generates
      // base env var names (e.g. POSTGRES_DSN), then replace with the actual
      // env var names we're reusing (e.g. POSTGRES_DSN or POSTGRES_DSN_1).
      let yaml = compileConnectorYAML(connectorDriver, values, {
        connectorInstanceName: connectorName,
        orderedProperties: schemaFields,
        fieldFilter: (p) => !p.internal,
        secretKeys: secretKeysList,
        stringKeys: stringKeysList,
        schema: schema ?? undefined,
      });

      // Replace generated env var refs with the actual (possibly suffixed) ones
      for (const key of secretKeysList) {
        const actual = existingEnvVars[key];
        if (!actual) continue;
        const base = makeEnvVarKey(
          connectorDriver.name as string,
          key,
          undefined,
          schema ?? undefined,
        );
        if (base !== actual) {
          yaml = yaml.replace(
            `{{ .env.${base} }}`,
            `{{ .env.${actual} }}`,
          );
        }
      }

      // Write connector YAML
      fileArtifact.updateEditorContent(yaml);
      await fileArtifact.saveLocalContent(true);
      formDirty = false;
    } catch (e) {
      console.error("Failed to save connector:", e);
    } finally {
      saving = false;
    }
  }

  function revert() {
    // Reload form from resource, discarding unsaved form changes
    if (schema && connectorName && schemaName) {
      loadFormFromResource(connectorName, schemaName);
    }
  }

  $: allErrorsStore = fileArtifact.getAllErrors(queryClient);
  $: allErrors = $allErrorsStore;
  $: errorMessage = allErrors[0]?.message;
  let errorDismissed = false;
  // Reset dismiss when error changes
  $: if (errorMessage) errorDismissed = false;
</script>

<div class="visual-connector">
  <div class="form-area">
    {#if schema}
      <div class="form-fields">
        <JSONSchemaFormRenderer
          {schema}
          step="connector"
          form={formStoreWrapper}
          errors={$errorsStore}
          onStringInputChange={handleStringInputChange}
          {handleFileUpload}
          iconMap={ICONS}
        />
      </div>
    {:else if driver}
      <div class="no-schema">
        <p class="text-sm text-fg-secondary">
          No visual editor available for driver "{driver}". Use the code editor
          instead.
        </p>
      </div>
    {:else}
      <div class="no-schema">
        <p class="text-sm text-fg-secondary">
          Add a <code>driver</code> field to your connector YAML to enable the visual
          editor.
        </p>
      </div>
    {/if}
  </div>

  {#if errorMessage && !errorDismissed}
    <div class="error-bar" role="status">
      <div class="flex gap-x-2 items-center flex-1 min-w-0">
        <CancelCircle className="text-destructive flex-none" />
        <span class="error-text">{errorMessage}</span>
      </div>
      <button
        class="dismiss-btn"
        aria-label="Dismiss error"
        on:click={() => (errorDismissed = true)}
      >
        <XIcon size="14px" />
      </button>
    </div>
  {/if}

  <footer class="save-bar">
    <div class="flex gap-x-3">
      <Button
        type="primary"
        disabled={!formDirty || saving}
        loading={saving}
        loadingCopy="Saving"
        onClick={save}
      >
        <Check size="14px" />
        Save
      </Button>

      <Button
        type="text"
        disabled={!formDirty || saving}
        onClick={revert}
      >
        <UndoIcon size="14px" />
        Revert changes
      </Button>
    </div>
  </footer>
</div>

<style lang="postcss">
  .visual-connector {
    @apply size-full flex flex-col bg-surface-background overflow-hidden;
  }

  .form-area {
    @apply flex-1 overflow-y-auto p-6 flex flex-col gap-y-4;
  }

  .form-fields {
    @apply max-w-lg flex flex-col gap-y-1;
  }

  .no-schema {
    @apply py-12 text-center;
  }

  .error-bar {
    @apply flex items-center gap-x-2 border-t border-destructive bg-destructive/15 px-4 py-3 text-sm text-fg-primary;
  }

  .error-text {
    @apply truncate;
  }

  .dismiss-btn {
    @apply flex-none p-1 rounded text-fg-muted;
  }

  .dismiss-btn:hover {
    @apply text-fg-primary bg-surface-hover;
  }

  .save-bar {
    @apply flex items-center px-6 py-3 border-t bg-surface-subtle;
  }
</style>
