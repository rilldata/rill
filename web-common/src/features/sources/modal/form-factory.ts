import { superForm, defaults, type SuperValidated } from "sveltekit-superforms";
import { yup } from "sveltekit-superforms/adapters";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { AddDataFormType } from "./types";
import type { ConnectorHandler } from "./connector-handlers";
import { dsnSchema } from "./yupSchemas";
import { getConnectorHandler } from "./connector-handlers";

export interface SuperFormState {
  form: SuperValidated<any>;
  errors: any;
  enhance: any;
  tainted: any;
  submit: any;
  submitting: boolean;
}

export interface FormCreationOptions {
  connector: V1ConnectorDriver;
  formType: AddDataFormType;
  formMode: "params" | "dsn";
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

/**
 * Create a superform instance for a connector
 */
export function createForm(options: FormCreationOptions): SuperFormState {
  const { connector, formType, formMode, onSubmit } = options;
  const handler = getConnectorHandler(connector);

  let schema: any;
  let initialValues: Record<string, unknown>;
  let formId: string;

  if (formMode === "dsn") {
    schema = yup(dsnSchema);
    initialValues = defaults(schema);
    formId = handler.getFormId(connector, "dsn");
  } else {
    schema = yup(handler.getValidationSchema());
    initialValues = handler.getInitialValues(connector, formType);
    formId = handler.getFormId(connector, "params");
  }

  const { form, errors, enhance, tainted, submit, submitting } = superForm(
    initialValues,
    {
      SPA: true,
      validators: schema,
      onUpdate: async (event) => {
        if (!event.form.valid) return;

        try {
          await onSubmit(event.form.data);
        } catch (error) {
          // Error handling will be done by the parent component
          throw error;
        }
      },
      resetForm: false,
    },
  );

  return {
    form: form as any,
    errors,
    enhance,
    tainted,
    submit,
    submitting: submitting as any,
  };
}

/**
 * Create multiple forms for a connector (params and dsn if available)
 */
export function createConnectorForms(
  connector: V1ConnectorDriver,
  formType: AddDataFormType,
  onSubmit: (values: Record<string, unknown>) => Promise<void>,
): {
  paramsForm: SuperFormState | null;
  dsnForm: SuperFormState | null;
} {
  const handler = getConnectorHandler(connector);

  // Always create params form
  const paramsForm = createForm({
    connector,
    formType,
    formMode: "params",
    onSubmit,
  });

  // Create DSN form if available
  let dsnForm: SuperFormState | null = null;
  if (handler.hasDsnFormOption(connector)) {
    dsnForm = createForm({
      connector,
      formType,
      formMode: "dsn",
      onSubmit,
    });
  }

  return {
    paramsForm,
    dsnForm,
  };
}

/**
 * Create ClickHouse-specific forms
 */
export function createClickHouseForms(
  connector: V1ConnectorDriver,
  formType: AddDataFormType,
  onSubmit: (values: Record<string, unknown>) => Promise<void>,
): {
  paramsForm: SuperFormState;
  dsnForm: SuperFormState;
} {
  const handler = getConnectorHandler(connector);

  const paramsForm = createForm({
    connector,
    formType,
    formMode: "params",
    onSubmit,
  });

  const dsnForm = createForm({
    connector,
    formType,
    formMode: "dsn",
    onSubmit,
  });

  return {
    paramsForm,
    dsnForm,
  };
}

/**
 * Get the appropriate form based on connector and current tab
 */
export function getCurrentForm(
  connector: V1ConnectorDriver,
  connectionTab: "parameters" | "dsn",
  forms: {
    paramsForm: SuperFormState | null;
    dsnForm: SuperFormState | null;
  },
): SuperFormState | null {
  const handler = getConnectorHandler(connector);

  if (handler.hasOnlyDsn(connector) || connectionTab === "dsn") {
    return forms.dsnForm;
  } else {
    return forms.paramsForm;
  }
}

/**
 * Get the appropriate form ID based on connector and current tab
 */
export function getCurrentFormId(
  connector: V1ConnectorDriver,
  connectionTab: "parameters" | "dsn",
): string {
  const handler = getConnectorHandler(connector);

  if (handler.hasOnlyDsn(connector) || connectionTab === "dsn") {
    return handler.getFormId(connector, "dsn");
  } else {
    return handler.getFormId(connector, "params");
  }
}

/**
 * Check if submit button should be disabled
 */
export function isSubmitDisabled(
  connector: V1ConnectorDriver,
  connectionTab: "parameters" | "dsn",
  forms: {
    paramsForm: SuperFormState | null;
    dsnForm: SuperFormState | null;
  },
): boolean {
  const handler = getConnectorHandler(connector);
  const currentForm = getCurrentForm(connector, connectionTab, forms);

  if (!currentForm) return true;

  // Check if form is submitting
  if (currentForm.submitting) return true;

  // Check required fields
  const properties =
    handler.hasOnlyDsn(connector) || connectionTab === "dsn"
      ? handler.getDsnProperties(connector)
      : handler.getFilteredProperties(connector, "connector");

  for (const property of properties) {
    if (property.required) {
      const key = String(property.key);
      const value = currentForm.form[key];
      if (isEmpty(value) || currentForm.errors[key]?.length) {
        return true;
      }
    }
  }

  return false;
}

/**
 * Helper function to check if a value is empty
 */
function isEmpty(value: unknown): boolean {
  if (value === null || value === undefined) return true;
  if (typeof value === "string") return value.trim() === "";
  if (typeof value === "number") return false;
  if (Array.isArray(value)) return value.length === 0;
  if (typeof value === "object") return Object.keys(value).length === 0;
  return false;
}
