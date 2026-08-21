import { CloudFeedbackState } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import { createRuntimeServiceListAIFeedback } from "@rilldata/web-common/runtime-client";
import { createLocalServiceListProjectAIFeedback } from "@rilldata/web-common/runtime-client/local-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, readable, type Readable } from "svelte/store";

// createOpenFeedbackCount returns the total number of open feedback items across
// the given runtime stores, plus (web-local only) the project's cloud deployment
// via the LocalService bridge. Powers the notification badge on the AI feedback
// navigation entry; the counts refresh through the same query invalidations as
// the inbox lists.
export function createOpenFeedbackCount(
  clients: (RuntimeClient | null)[],
  includeLocalServiceBridge: boolean,
): Readable<number> {
  const counts: Readable<number>[] = [];
  for (const client of clients) {
    if (!client) continue;
    counts.push(
      derived(
        createRuntimeServiceListAIFeedback(client, { status: "open" }),
        ($q) => $q.data?.feedback?.length ?? 0,
      ),
    );
  }
  if (includeLocalServiceBridge) {
    counts.push(
      derived(createLocalServiceListProjectAIFeedback({ status: "open" }), ($q) =>
        $q.data?.cloudState === CloudFeedbackState.OK
          ? ($q.data?.feedback?.length ?? 0)
          : 0,
      ),
    );
  }
  if (counts.length === 0) return readable(0);
  return derived(counts, (values) => values.reduce((sum, n) => sum + n, 0));
}
