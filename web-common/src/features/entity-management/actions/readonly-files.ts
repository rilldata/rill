import type { Snippet } from "svelte";

const PROTECTED_DIRECTORIES: ReadonlyFileMatcher[] = [
  {
    id: "tmp",
    matcher: /^\/tmp(\/|$)/,
  },
  {
    id: "git",
    matcher: /^\/.git(\/|$)/,
  },
];
const PROTECTED_FILES: ReadonlyFileMatcher[] = [
  {
    id: "/rill.yaml",
    matcher: /^\/rill.yaml$/,
    allowFileEdit: true,
  },
];

export class ReadonlyFiles {
  private readonly readonlyFiles = new Map<string, ReadonlyFileMatcher>(
    PROTECTED_FILES.map((matcher) => [matcher.id, matcher]),
  );
  private readonly readonlyDirs = new Map<string, ReadonlyFileMatcher>(
    PROTECTED_DIRECTORIES.map((matcher) => [matcher.id, matcher]),
  );

  public addReadonly(matcher: ReadonlyFileMatcher) {
    this.readonlyFiles.set(matcher.id, matcher);
  }

  public match(path: string) {
    return this.readonlyFiles
      .values()
      .find((matcher) => matcher.matcher.test(path));
  }

  public matchDir(path: string) {
    return this.readonlyDirs
      .values()
      .find((matcher) => matcher.matcher.test(path));
  }
}

type ReadonlyFileMatcher = {
  id: string;
  matcher: RegExp;
  messageSnippet?: Snippet;
  // Allows file edit but disables rename/delete
  allowFileEdit?: boolean;
};

export const readonlyFiles = new ReadonlyFiles();
