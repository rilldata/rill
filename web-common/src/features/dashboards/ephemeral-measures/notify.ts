import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

/**
 * Confirms a saved or deleted ephemeral measure with a toast. Callers emit it
 * once the change is persisted: right away for the explore state, after the
 * file write for a canvas component.
 */
export function notifyEphemeralMeasureChange(
  change: "created" | "updated" | "deleted",
  displayName: string,
) {
  const message = {
    created: m.dashboard_pivot_ephemeral_created,
    updated: m.dashboard_pivot_ephemeral_updated,
    deleted: m.dashboard_pivot_ephemeral_deleted,
  }[change]({ name: displayName });
  eventBus.emit("notification", { type: "success", message });
}
