import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import {
  queryServiceConvertExpressionToMetricsSQL,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

/**
 * Converts the filter applied in an explore to the Metrics SQL that `defaults.filters` holds.
 *
 * Resolves to an empty string when no filter is applied, and to undefined when the conversion
 * failed. A failure is reported to the user, and the caller must not save the defaults then:
 * writing them without the filter would drop the default filter the file already has.
 */
export async function convertFilterToDefaultFilterSQL(
  client: RuntimeClient,
  expr: V1Expression | undefined,
): Promise<string | undefined> {
  if (!expr?.cond?.exprs?.length) return "";

  try {
    const resp = await queryServiceConvertExpressionToMetricsSQL(client, {
      // TODO: the API type is the proto message but takes its JSON form. See canvas-entity.ts.
      expression: expr as any,
    });
    return resp.sql ?? "";
  } catch {
    eventBus.emit("notification", {
      type: "error",
      message: m.explore_save_default_filter_error(),
    });
    return undefined;
  }
}
