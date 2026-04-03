import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import {
  getMessage,
  MessageSelectors,
} from "@rilldata/web-common/features/chat/core/messages/message-selectors.ts";
import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";

// =============================================================================
// BACKEND TYPES (mirror runtime/ai request_connector_fields tool definitions)
// =============================================================================

/** Arguments for the request_connector_fields tool call */
export interface RequestConnectorFieldsCallData {
  driver: string;
  missing_fields: string[];
  message?: string;
  connector_path?: string;
}

/** Result from the request_connector_fields tool */
export interface RequestConnectorFieldsResultData {
  driver: string;
  missing_fields: string[];
  message?: string;
  connector_path?: string;
}

/**
 * Block for the `request_connector_fields` tool: shows the tool call/result and
 * reserves space for connector credential forms (filled in later).
 */
export type RequestConnectorFieldsBlock = {
  type: "request-connector-fields-block";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  llmMessage: string | undefined;
  schemaName: string;
  connectorPath: string;
  schema: MultiStepFormSchema;
  filteredSchema: MultiStepFormSchema;
  hasSubmitted: boolean;
};

export function createRequestConnectorFieldsBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
  allMessages: V1Message[],
): RequestConnectorFieldsBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  try {
    const callData = JSON.parse(
      message.contentData || "{}",
    ) as RequestConnectorFieldsCallData;
    if (!callData.driver?.trim()) return null;

    const resultData = JSON.parse(
      resultMessage.contentData || "{}",
    ) as RequestConnectorFieldsResultData;
    if (!resultData.driver?.trim() || !resultData.missing_fields?.length) {
      return null;
    }

    const schema = getConnectorSchema(resultData.driver);
    if (!schema?.properties) return null;
    const filteredSchema = structuredClone(schema);
    filteredSchema.properties = Object.fromEntries(
      Object.keys(schema.properties)
        .filter((k) => resultData.missing_fields.includes(k))
        .map((k) => [k, schema.properties![k]]),
    );

    const resultIndex = allMessages.findIndex((m) => m.id === resultMessage.id);
    if (resultIndex === -1) return null;
    const userMessageAfterResult = getMessage(
      allMessages,
      [MessageSelectors.ByRoleName("user")],
      resultIndex,
    );

    return {
      type: "request-connector-fields-block",
      id: `request-connector-fields-${message.id}`,
      message,
      resultMessage,
      llmMessage: resultData.message,
      schemaName: resultData.driver,
      schema,
      filteredSchema,
      connectorPath: resultData.connector_path
        ? addLeadingSlash(resultData.connector_path)
        : "",
      hasSubmitted: !!userMessageAfterResult,
    };
  } catch {
    return null;
  }
}
