import { isAdminServerQuery } from "@rilldata/web-admin/client/utils";
import { errorStore } from "@rilldata/web-admin/components/errors/error-store";
import { createUserFacingError } from "@rilldata/web-admin/components/errors/user-facing-errors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import type { Query, QueryClient } from "@tanstack/svelte-query";

export const AdminNetworkErrorMessage = "Network Error";

const AdminNetworkBannerID = "admin-network";
const AdminNetworkBannerPriority = 0;

type QueryRefetcher = Pick<QueryClient, "refetchQueries">;

let adminNetworkErrorActive = false;

export function isNetworkError(error: unknown): boolean {
  return error instanceof Error && error.message === AdminNetworkErrorMessage;
}

export function handleAdminServerNetworkError(
  error: unknown,
  query: Query,
  queryClient: QueryRefetcher,
): boolean {
  if (!isNetworkError(error) || !isAdminServerQuery(query)) return false;

  adminNetworkErrorActive = true;

  if (queryHasCachedData(query)) {
    showAdminNetworkBanner(queryClient);
  } else {
    showAdminNetworkErrorPage();
  }

  return true;
}

export function handleAdminServerQuerySuccess(query: Query): void {
  if (!adminNetworkErrorActive || !isAdminServerQuery(query)) return;
  clearAdminNetworkErrorState();
}

export function recoverFromAdminNetworkError(
  queryClient: QueryRefetcher,
): Promise<unknown> {
  if (!adminNetworkErrorActive) return Promise.resolve();

  clearAdminNetworkErrorState();
  return queryClient.refetchQueries({
    type: "active",
    predicate: isAdminServerQuery,
  });
}

export function registerAdminNetworkRecoveryListeners(
  queryClient: QueryRefetcher,
): () => void {
  if (typeof window === "undefined") return () => {};

  const recover = () => {
    void recoverFromAdminNetworkError(queryClient);
  };
  const recoverWhenVisible = () => {
    if (document.visibilityState !== "hidden") recover();
  };

  window.addEventListener("online", recover);
  window.addEventListener("focus", recoverWhenVisible);
  document.addEventListener("visibilitychange", recoverWhenVisible);

  return () => {
    window.removeEventListener("online", recover);
    window.removeEventListener("focus", recoverWhenVisible);
    document.removeEventListener("visibilitychange", recoverWhenVisible);
  };
}

export function clearAdminNetworkErrorState(): void {
  adminNetworkErrorActive = false;
  errorStore.reset();
  eventBus.emit("remove-banner", AdminNetworkBannerID);
}

function queryHasCachedData(query: Query): boolean {
  return query.state.data !== undefined;
}

function showAdminNetworkErrorPage(): void {
  errorStore.set(createUserFacingError(null, AdminNetworkErrorMessage));
}

function showAdminNetworkBanner(queryClient: QueryRefetcher): void {
  eventBus.emit("add-banner", {
    id: AdminNetworkBannerID,
    priority: AdminNetworkBannerPriority,
    message: {
      type: "warning",
      iconType: "alert",
      message:
        "Connection to Rill Cloud was interrupted. Showing cached data while we reconnect.",
      cta: {
        type: "button",
        text: "Retry now",
        async onClick() {
          await recoverFromAdminNetworkError(queryClient);
        },
      },
    },
  });
}
