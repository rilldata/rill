import { getContext, setContext } from "svelte";
import type { Snippet } from "svelte";

export type ReadonlyMatcher = {
  matcher: RegExp;
  messageSnippet?: Snippet;
  // Allows file edit but disables rename/delete
  allowFileEdit?: boolean;
};

const PROTECTED_FILES: ReadonlyMatcher[] = [
  { matcher: /^\/rill\.yaml$/, allowFileEdit: true },
];

const PROTECTED_DIRECTORIES: RegExp[] = [/^\/tmp(\/|$)/, /^\/\.git(\/|$)/];

// Route subtrees publish additional readonly matchers via setContext. Scope is
// natural: when the layout unmounts, the entries leave with it. The `.env`
// readonly rule only applies inside the cloud edit layout, so it lives there.
const READONLY_FILES_CONTEXT = Symbol("readonly-files");

export function getReadonlyExtras(): ReadonlyMatcher[] {
  return (
    getContext<ReadonlyMatcher[] | undefined>(READONLY_FILES_CONTEXT) ?? []
  );
}

export function setReadonlyExtras(matchers: ReadonlyMatcher[]) {
  setContext(READONLY_FILES_CONTEXT, matchers);
}

export function matchReadonlyFile(
  path: string,
  extras: ReadonlyMatcher[] = [],
): ReadonlyMatcher | undefined {
  for (const m of PROTECTED_FILES) if (m.matcher.test(path)) return m;
  for (const m of extras) if (m.matcher.test(path)) return m;
  return undefined;
}

export function matchReadonlyDir(path: string): boolean {
  return PROTECTED_DIRECTORIES.some((re) => re.test(path));
}
