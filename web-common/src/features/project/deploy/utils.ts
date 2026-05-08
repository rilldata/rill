const targetDashboardKey = "rill:app:targetDashboard";
export const TargetDashboardUrlParam = "target_dashboard";
export const PreCommitShaUrlParam = "pre_commit_sha";

/**
 * Sets the target dashboard name from url to session storage if present.
 * This prevents having to pass through the param during deploy flow. We could be going through different routes there.
 */
export function maybeSetTargetDashboard(url: URL) {
  const targetDashboard = url.searchParams.get(TargetDashboardUrlParam);
  if (!targetDashboard) {
    // Remove item to ensure a stale dashboard name is not used.
    sessionStorage.removeItem(targetDashboardKey);
  } else {
    sessionStorage.setItem(targetDashboardKey, targetDashboard);
  }
}

export function getTargetDashboard() {
  const url = new URL(window.location.href);
  return (
    sessionStorage.getItem(targetDashboardKey) ||
    url.searchParams.get(TargetDashboardUrlParam)
  );
}
