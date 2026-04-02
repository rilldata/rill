// =============================================================================
// BACKEND TYPES (mirror runtime/ai request_connector_fields tool definitions)
// =============================================================================

/** Arguments for the request_connector_fields tool call */
export interface RequestConnectorFieldsCallData {
  driver: string;
  entered_fields?: Record<string, string>;
  missing_fields: string[];
  message?: string;
  resource_name?: string;
}

/** Result from the request_connector_fields tool */
export interface RequestConnectorFieldsResultData {
  driver: string;
  entered_fields?: Record<string, string>;
  missing_fields: string[];
  message?: string;
  resource_name?: string;
}
