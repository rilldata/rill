import { getContext, setContext } from "svelte";
import type { Snippet } from "svelte";
import {
  extractFileName,
  splitFolderAndFileName,
} from "@rilldata/web-common/features/entity-management/file-path-utils.ts";

export type ReadonlyMatcher = {
  matcher: RegExp;
  messageSnippet?: Snippet;
  // Allows file edit but disables rename/delete
  // Currently rill.yaml falls under this
  allowFileEdit?: boolean;
};

const PROTECTED_FILES: ReadonlyMatcher[] = [
  { matcher: /rill\.yaml/, allowFileEdit: true },
];

const PROTECTED_DIRECTORIES: RegExp[] = [/^\/tmp(\/|$)/, /^\/\.git(\/|$)/];

// Route subtrees publish additional readonly matchers via setContext. Scope is
// natural: when the layout unmounts, the entries leave with it. The `.env`
// readonly rule only applies inside the cloud edit layout, so it lives there.
const READONLY_FILES_CONTEXT = Symbol("readonly-files");

export const ReadonlyEnvFilesRegex = /^\.([^.]+\.)?env$/;

export function getAdditionalReadonlyFiles(): ReadonlyMatcher[] {
  return (
    getContext<ReadonlyMatcher[] | undefined>(READONLY_FILES_CONTEXT) ?? []
  );
}

export function setAdditionalReadonlyFiles(matchers: ReadonlyMatcher[]) {
  setContext(READONLY_FILES_CONTEXT, matchers);
}

export function matchReadonlyFile(
  path: string,
  additionalReadonlyFiles: ReadonlyMatcher[] = [],
): ReadonlyMatcher | undefined {
  for (const m of PROTECTED_FILES) if (m.matcher.test(path)) return m;
  if (additionalReadonlyFiles.length === 0) return undefined;
  const [, fileName] = splitFolderAndFileName(path);
  for (const m of additionalReadonlyFiles)
    if (m.matcher.test(fileName)) return m;
  return undefined;
}

export function matchReadonlyDir(path: string): boolean {
  return PROTECTED_DIRECTORIES.some((re) => re.test(path));
}
