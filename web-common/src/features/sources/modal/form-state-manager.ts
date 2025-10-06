import { writable, get } from "svelte/store";
import type { SuperValidated } from "sveltekit-superforms";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { AddDataFormType, ConnectorType } from "./types";
import type { ClickHouseConnectorType } from "./constants";

export interface FormError {
  message: string;
  details?: string;
}

export interface FormState {
  connector: V1ConnectorDriver;
  formType: AddDataFormType;
  connectionTab: ConnectorType;
  clickhouseConnectorType: ClickHouseConnectorType;
  submitting: boolean;
  copied: boolean;
}

export interface SuperFormState {
  form: SuperValidated<any>;
  errors: any;
  enhance: any;
  tainted: any;
  submit: any;
  submitting: boolean;
}

export interface FormStates {
  params: SuperFormState | null;
  dsn: SuperFormState | null;
  clickhouse: SuperFormState | null;
}

export interface FormErrors {
  params: FormError | null;
  dsn: FormError | null;
  clickhouse: FormError | null;
}

export class FormStateManager {
  public readonly state = writable<FormState>();
  private _formStates: FormStates;
  private _errors: FormErrors;

  constructor(connector: V1ConnectorDriver, formType: AddDataFormType) {
    this.state.set({
      connector,
      formType,
      connectionTab: "parameters",
      clickhouseConnectorType: "self-hosted",
      submitting: false,
      copied: false,
    });

    this._formStates = {
      params: null,
      dsn: null,
      clickhouse: null,
    };

    this._errors = {
      params: null,
      dsn: null,
      clickhouse: null,
    };
  }

  // Getters
  get connector(): V1ConnectorDriver {
    return get(this.state).connector;
  }

  get formType(): AddDataFormType {
    return get(this.state).formType;
  }

  get connectionTab(): ConnectorType {
    return get(this.state).connectionTab;
  }

  get clickhouseConnectorType(): ClickHouseConnectorType {
    return get(this.state).clickhouseConnectorType;
  }

  get submitting(): boolean {
    return get(this.state).submitting;
  }

  get copied(): boolean {
    return get(this.state).copied;
  }

  get formStates(): FormStates {
    return this._formStates;
  }

  get errors(): FormErrors {
    return this._errors;
  }

  // Setters
  setConnectionTab(tab: ConnectorType): void {
    this.state.update((state) => ({ ...state, connectionTab: tab }));
  }

  setClickhouseConnectorType(type: ClickHouseConnectorType): void {
    this.state.update((state) => {
      const newState = { ...state, clickhouseConnectorType: type };
      // Reset connectionTab if switching to Rill-managed
      if (type === "rill-managed") {
        newState.connectionTab = "parameters";
      }
      return newState;
    });
  }

  setSubmitting(submitting: boolean): void {
    this.state.update((state) => ({ ...state, submitting }));
  }

  setCopied(copied: boolean): void {
    this.state.update((state) => ({ ...state, copied }));
  }

  // Form state management
  setFormState(
    type: "params" | "dsn" | "clickhouse",
    formState: SuperFormState,
  ): void {
    this._formStates[type] = formState;
  }

  getFormState(type: "params" | "dsn" | "clickhouse"): SuperFormState | null {
    return this._formStates[type];
  }

  // Error management
  setError(
    type: "params" | "dsn" | "clickhouse",
    error: FormError | null,
  ): void {
    this._errors[type] = error;
  }

  getError(type: "params" | "dsn" | "clickhouse"): FormError | null {
    return this._errors[type];
  }

  clearAllErrors(): void {
    this._errors = {
      params: null,
      dsn: null,
      clickhouse: null,
    };
  }

  // Computed properties
  get isSourceForm(): boolean {
    return get(this.state).formType === "source";
  }

  get isConnectorForm(): boolean {
    return get(this.state).formType === "connector";
  }

  get hasDsnFormOption(): boolean {
    const currentState = get(this.state);
    return (
      this.isConnectorForm &&
      Boolean(
        currentState.connector.configProperties?.some(
          (property) => property.key === "dsn",
        ),
      ) &&
      Boolean(
        currentState.connector.configProperties?.some(
          (property) => property.key !== "dsn",
        ),
      )
    );
  }

  get hasOnlyDsn(): boolean {
    const currentState = get(this.state);
    return (
      this.isConnectorForm &&
      Boolean(
        currentState.connector.configProperties?.some(
          (property) => property.key === "dsn",
        ),
      ) &&
      !Boolean(
        currentState.connector.configProperties?.some(
          (property) => property.key !== "dsn",
        ),
      )
    );
  }

  get isClickhouseConnector(): boolean {
    return get(this.state).connector.name === "clickhouse";
  }

  // Get current active form based on connector and tab
  getCurrentFormState(): SuperFormState | null {
    const currentState = get(this.state);
    if (this.isClickhouseConnector) {
      return this._formStates.clickhouse;
    } else if (this.hasOnlyDsn || currentState.connectionTab === "dsn") {
      return this._formStates.dsn;
    } else {
      return this._formStates.params;
    }
  }

  // Get current active error based on connector and tab
  getCurrentError(): FormError | null {
    const currentState = get(this.state);
    if (this.isClickhouseConnector) {
      return this._errors.clickhouse;
    } else if (this.hasOnlyDsn || currentState.connectionTab === "dsn") {
      return this._errors.dsn;
    } else {
      return this._errors.params;
    }
  }

  // Get current form ID
  getCurrentFormId(): string {
    const currentState = get(this.state);
    const connectorName = currentState.connector.name;

    if (this.isClickhouseConnector) {
      return currentState.connectionTab === "parameters"
        ? `add-clickhouse-data-${connectorName}-form`
        : `add-clickhouse-data-${connectorName}-dsn-form`;
    } else if (this.hasOnlyDsn || currentState.connectionTab === "dsn") {
      return `add-data-${connectorName}-dsn-form`;
    } else {
      return `add-data-${connectorName}-form`;
    }
  }

  // Get submit button text
  getSubmitButtonText(): string {
    const currentState = get(this.state);
    if (this.isClickhouseConnector) {
      if (currentState.clickhouseConnectorType === "rill-managed") {
        return currentState.submitting ? "Connecting..." : "Connect";
      } else {
        return currentState.submitting
          ? "Testing connection..."
          : "Test and Connect";
      }
    } else if (this.isConnectorForm) {
      return currentState.submitting
        ? "Testing connection..."
        : "Test and Connect";
    } else {
      return "Test and Add data";
    }
  }

  // Get loading copy text
  getLoadingCopyText(): string {
    if (this.isClickhouseConnector) {
      return "Connecting...";
    } else {
      return "Testing connection...";
    }
  }
}
