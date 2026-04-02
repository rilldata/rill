import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";
import type {
  RequestConnectorFieldsCallData,
  RequestConnectorFieldsResultData,
} from "./request-connector-fields-data.ts";
import type { MultiStepFormSchema } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { getConnectorSchema } from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";

export type {
  RequestConnectorFieldsCallData,
  RequestConnectorFieldsResultData,
} from "./request-connector-fields-data.ts";

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
  schema: MultiStepFormSchema;
  enteredFields: Record<string, any>;
};

export function createRequestConnectorFieldsBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
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

    const schema = structuredClone(getConnectorSchema(resultData.driver));
    if (!schema?.properties) return null;
    schema.properties = Object.fromEntries(
      Object.keys(schema.properties)
        .filter((k) => resultData.missing_fields.includes(k))
        .map((k) => [k, schema.properties![k]]),
    );

    return {
      type: "request-connector-fields-block",
      id: `request-connector-fields-${message.id}`,
      message,
      resultMessage,
      llmMessage: resultData.message,
      schemaName: resultData.driver,
      schema,
      enteredFields: resultData.entered_fields ?? {},
    };
  } catch {
    return null;
  }
}
