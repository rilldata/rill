const deployingNameKey = "rill:app:deployingName";

export function maybeSetDeployingName(url: URL) {
  const deployingName = url.searchParams.get("deploying_name");
  if (!deployingName) return;
  sessionStorage.setItem(deployingNameKey, deployingName);
}

export function getDeployingName() {
  return sessionStorage.getItem(deployingNameKey);
}
