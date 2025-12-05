import type { V1Message } from "@rilldata/web-common/runtime-client";
import {
  MessageContentType,
  type WriteFileCallData,
  type WriteFileResultData,
} from "../../types";

/**
 * File diff block representation.
 * Contains a diff visualization extracted from a write_file tool call.
 */
export type FileDiffBlock = {
  type: "file-diff";
  id: string;
  message: V1Message;
  filePath: string;
  diff: string;
  isNewFile: boolean;
};

/**
 * Creates a file diff block from a write_file tool call message.
 * Returns null if the data is invalid or the result indicates an error.
 */
export function createFileDiffBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): FileDiffBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  try {
    const callData: WriteFileCallData = JSON.parse(message.contentData || "{}");
    const filePath = callData.path || "";
    if (!filePath) return null;

    const resultData: WriteFileResultData = JSON.parse(
      resultMessage.contentData || "{}",
    );

    return {
      type: "file-diff",
      id: `file-diff-${message.id}`,
      message,
      filePath,
      diff: resultData.diff || "",
      isNewFile: resultData.is_new_file || false,
    };
  } catch {
    return null;
  }
}
