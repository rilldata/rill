const deployingDashboardKey = "rill:app:deployingDashboard";
export const DeployingDashboardUrlParam = "deploying_dashboard";

/**
 * Sets the deploying dashboard name from url to session storage if present.
 * This prevents having to pass through the param during deploy flow. We could be going through different routes there.
 */
export function maybeSetDeployingDashboard(url: URL) {
  const deployingDashboard = url.searchParams.get(DeployingDashboardUrlParam);
  if (!deployingDashboard) {
    // Remove item to ensure a stale dashboard name is not used.
    sessionStorage.removeItem(deployingDashboardKey);
  } else {
    sessionStorage.setItem(deployingDashboardKey, deployingDashboard);
  }
}

export function getDeployingDashboard() {
  return sessionStorage.getItem(deployingDashboardKey);
}
