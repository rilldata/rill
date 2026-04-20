/**
 * Returns the default project files for a new project.
 * After a bit of debate on using runtime methods like InitEmpty directly or running `rill init`
 * or waiting for deployment is finished in UI and calling UnpackEmpty API, this is the best option right now.
 * We might revisit this in the future.
 */
export function getProjectInitFiles(
  displayName: string,
): Record<string, string> {
  return {
    "rill.yaml": `compiler: rillv1

display_name: ${displayName}

# The project's default OLAP connector.
# Learn more: https://docs.rilldata.com/reference/olap-engines
olap_connector: duckdb

# These are example mock users to test your security policies.
# Learn more: https://docs.rilldata.com/developers/build/rill-project-file#test-access-policies-in-rill-developer
mock_users:
- email: john@yourcompany.com
- email: jane@partnercompany.com
`,
    ".gitignore": ".DS_Store\n\n# Rill\n.env\ntmp\n",
  };
}
