#!/usr/bin/env node
// Asserts route-tree parity between Rill Developer (web-local) and the
// cloud editing surface (web-admin/[organization]/[project]/-/edit/).
//
// Both trees render the same shared components from web-common and rely on
// `editorRoutePrefix` (web-common/src/layout/navigation/editor-routing.ts) so
// shared navigation works in either context. This check fails CI when a
// route exists on one side but not the other, with an explicit allowlist for
// intentional exclusions. Pattern mirrors scripts/tsc-with-whitelist.sh.

import { existsSync, readdirSync, statSync } from "node:fs";
import { join, relative, sep } from "node:path";
import { fileURLToPath } from "node:url";

const ROUTE_FILE_PATTERN = /^\+(page|layout|server)\.(svelte|ts)$/;

const REPO_ROOT = fileURLToPath(new URL("..", import.meta.url));

const LOCAL_ROOTS = [
  "web-local/src/routes/(application)/(workspace)",
  "web-local/src/routes/(viz)",
  "web-local/src/routes/(misc)",
];
const ADMIN_ROOT = "web-admin/src/routes/[organization]/[project]/-/edit";

// Logical paths that exist only in web-local by design. Keep a short reason
// comment on each entry so a future contributor can judge whether it's still
// load-bearing.
const LOCAL_ONLY_ALLOWLIST = [
  // Citation URL routes
  // TODO: ensure citations within the edit session get routed to the developer
  // preview dashboards _within_ the edit session, not to the branch preview 
  // dashboards _outside_ the edit session.
  "/-/ai/[conversationId]/message/[messageId]/+layout.ts",
  "/-/ai/[conversationId]/message/[messageId]/-/open/+page.ts",
  "/-/open-query/+page.ts",

  // Backcompat redirect /dashboard/foo → /explore/foo. Cloud never exposed
  // /dashboard/[name] URLs, so there's nothing to redirect from. Permanent
  // local-only.
  "/dashboard/[name]/+page.ts",

  // Deploy related stuff is not on admin
  "/deploy/+layout.svelte",
  "/deploy/+page.svelte",
  "/deploy/+page.ts",
  "/deploy/organization/create/+page.svelte",
  "/deploy/organization/select/+page.svelte",
  "/deploy/project/create/+page.svelte",
  "/deploy/project/create/+page.ts",
  "/deploy/project/github/+page.svelte",
  "/deploy/project/github/+page.ts",
  "/deploy/project/select/+page.svelte",
  "/deploy/project/update/+page.svelte",
  "/deploy/project/update/+page.ts",
];

const ADMIN_ONLY_ALLOWLIST = [
  // We have a layout at the root on rill-dev, not under subpath like (application)/(workspace)/ or (viz)/
  "/+layout.ts",
];

function walkRoutes(absRoot) {
  const results = [];
  if (!existsSync(absRoot)) return results;
  const stack = [absRoot];
  while (stack.length > 0) {
    const dir = stack.pop();
    for (const entry of readdirSync(dir)) {
      const full = join(dir, entry);
      const stats = statSync(full);
      if (stats.isDirectory()) {
        stack.push(full);
      } else if (ROUTE_FILE_PATTERN.test(entry)) {
        const logical = "/" + relative(absRoot, full).split(sep).join("/");
        results.push(logical);
      }
    }
  }
  return results;
}

function collect(roots) {
  const set = new Set();
  for (const root of roots) {
    const abs = join(REPO_ROOT, root);
    if (!existsSync(abs)) {
      console.error(`ERROR: route root not found: ${root}`);
      process.exit(2);
    }
    for (const path of walkRoutes(abs)) set.add(path);
  }
  return set;
}

function diff(a, b, allowlist) {
  const allowed = new Set(allowlist);
  const missing = [];
  for (const path of a) {
    if (!b.has(path) && !allowed.has(path)) missing.push(path);
  }
  missing.sort();
  return missing;
}

function staleAllowlistEntries(allowlist, shouldExistIn) {
  return allowlist.filter((p) => !shouldExistIn.has(p)).sort();
}

function main() {
  const localRoutes = collect(LOCAL_ROOTS);
  const adminRoutes = collect([ADMIN_ROOT]);

  const missingInAdmin = diff(localRoutes, adminRoutes, LOCAL_ONLY_ALLOWLIST);
  const missingInLocal = diff(adminRoutes, localRoutes, ADMIN_ONLY_ALLOWLIST);

  // An allowlist entry must point at a route that still exists on the
  // "allowed" side — otherwise the allowlist is rotting.
  const staleLocalAllowlist = staleAllowlistEntries(
    LOCAL_ONLY_ALLOWLIST,
    localRoutes,
  );
  const staleAdminAllowlist = staleAllowlistEntries(
    ADMIN_ONLY_ALLOWLIST,
    adminRoutes,
  );

  let failed = false;

  if (missingInAdmin.length > 0) {
    failed = true;
    console.error(
      `\nRoutes present in web-local but missing in web-admin/-/edit/:`,
    );
    for (const p of missingInAdmin) console.error(`  ${p}`);
    console.error(
      `\nFix: either mirror each route under ${ADMIN_ROOT}/, or add the path to LOCAL_ONLY_ALLOWLIST in scripts/check-edit-route-parity.js with a reason.`,
    );
  }

  if (missingInLocal.length > 0) {
    failed = true;
    console.error(
      `\nRoutes present in web-admin/-/edit/ but missing in web-local:`,
    );
    for (const p of missingInLocal) console.error(`  ${p}`);
    console.error(
      `\nFix: either mirror each route under web-local/src/routes/(application)/(workspace)/ or (viz)/, or add the path to ADMIN_ONLY_ALLOWLIST in scripts/check-edit-route-parity.js with a reason.`,
    );
  }

  if (staleLocalAllowlist.length > 0) {
    failed = true;
    console.error(
      `\nLOCAL_ONLY_ALLOWLIST entries no longer correspond to a real file in web-local:`,
    );
    for (const p of staleLocalAllowlist) console.error(`  ${p}`);
    console.error(
      `\nFix: remove the stale entries from LOCAL_ONLY_ALLOWLIST in scripts/check-edit-route-parity.js.`,
    );
  }

  if (staleAdminAllowlist.length > 0) {
    failed = true;
    console.error(
      `\nADMIN_ONLY_ALLOWLIST entries no longer correspond to a real file in web-admin/-/edit/:`,
    );
    for (const p of staleAdminAllowlist) console.error(`  ${p}`);
    console.error(
      `\nFix: remove the stale entries from ADMIN_ONLY_ALLOWLIST in scripts/check-edit-route-parity.js.`,
    );
  }

  if (failed) {
    process.exit(1);
  }

  console.log(
    `Edit route parity OK (${localRoutes.size} local, ${adminRoutes.size} admin).`,
  );
}

main();
