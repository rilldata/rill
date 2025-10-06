import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type * as yup from "yup";
import type { AddDataFormType } from "./types";

export interface ConnectorProperty {
  key?: string;
  displayName?: string;
  placeholder?: string;
  hint?: string;
  secret?: boolean;
  required?: boolean;
  noPrompt?: boolean;
  type?: string;
}

export interface ConnectorFormConfig {
  properties: ConnectorProperty[];
  validationSchema: yup.ObjectSchema<any>;
  initialValues: Record<string, unknown>;
  formId: string;
}

export interface ConnectorHandler {
  /**
   * Get the connector name this handler manages
   */
  getConnectorName(): string;

  /**
   * Check if this handler supports the given connector
   */
  supports(connector: V1ConnectorDriver): boolean;

  /**
   * Get properties for the connector based on form type
   */
  getProperties(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
  ): ConnectorProperty[];

  /**
   * Get filtered properties (excluding noPrompt, etc.)
   */
  getFilteredProperties(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
  ): ConnectorProperty[];

  /**
   * Get validation schema for the connector
   */
  getValidationSchema(): yup.ObjectSchema<any>;

  /**
   * Get initial form values
   */
  getInitialValues(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
  ): Record<string, unknown>;

  /**
   * Get form ID for the connector
   */
  getFormId(connector: V1ConnectorDriver, formType: "params" | "dsn"): string;

  /**
   * Check if connector has DSN form option
   */
  hasDsnFormOption(connector: V1ConnectorDriver): boolean;

  /**
   * Check if connector only has DSN (no tabs)
   */
  hasOnlyDsn(connector: V1ConnectorDriver): boolean;

  /**
   * Get DSN properties if available
   */
  getDsnProperties(connector: V1ConnectorDriver): ConnectorProperty[];

  /**
   * Handle form submission
   */
  handleSubmit(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): Promise<void>;

  /**
   * Get YAML preview for the connector configuration
   */
  getYamlPreview(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): string;

  /**
   * Handle special connector-specific logic
   */
  handleSpecialLogic?(
    connector: V1ConnectorDriver,
    formType: AddDataFormType,
    values: Record<string, unknown>,
  ): Record<string, unknown>;
}

export class ConnectorHandlerRegistry {
  private handlers: Map<string, ConnectorHandler> = new Map();

  /**
   * Register a connector handler
   */
  register(handler: ConnectorHandler): void {
    this.handlers.set(handler.getConnectorName(), handler);
  }

  /**
   * Get a handler for a specific connector
   */
  getHandler(connectorName: string): ConnectorHandler | null {
    return this.handlers.get(connectorName) || null;
  }

  /**
   * Get a handler for a connector driver
   */
  getHandlerForConnector(
    connector: V1ConnectorDriver,
  ): ConnectorHandler | null {
    return this.getHandler(connector.name || "");
  }

  /**
   * Check if a handler exists for a connector
   */
  hasHandler(connectorName: string): boolean {
    return this.handlers.has(connectorName);
  }

  /**
   * Get all registered connector names
   */
  getRegisteredConnectors(): string[] {
    return Array.from(this.handlers.keys());
  }

  /**
   * Unregister a handler
   */
  unregister(connectorName: string): boolean {
    return this.handlers.delete(connectorName);
  }

  /**
   * Clear all handlers
   */
  clear(): void {
    this.handlers.clear();
  }
}

// Global registry instance
export const connectorHandlerRegistry = new ConnectorHandlerRegistry();

/**
 * Helper function to get handler for a connector
 */
export function getConnectorHandler(
  connector: V1ConnectorDriver,
): ConnectorHandler {
  const handler = connectorHandlerRegistry.getHandlerForConnector(connector);
  if (!handler) {
    throw new Error(`No handler found for connector: ${connector.name}`);
  }
  return handler;
}

/**
 * Helper function to check if connector has a handler
 */
export function hasConnectorHandler(connector: V1ConnectorDriver): boolean {
  return connectorHandlerRegistry.hasHandler(connector.name || "");
}
