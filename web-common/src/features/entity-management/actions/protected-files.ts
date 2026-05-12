import picomatch from "picomatch";
import type { Snippet } from "svelte";
import { isCloudRuntimeEditEnvironment } from "../edit-environment.ts";

// Two distinct kinds of path protections:
//
// - managed: file content is hidden and not editable, managed from sources outside of the editor.
//   Renames and deletes are also blocked because you can't act on a file you can't edit.
//   `.env` on cloud is the canonical case.
//
// - pinned: content is editable, but the path is locked. The file can't be
//   renamed or deleted, and other files can't be renamed onto it. `rill.yaml`
//   is the canonical case.
//
// The two are surfaced as separate predicates so that call sites express what
// they actually need, rather than a generic "protected" flag with hidden flips.

// `dot: true` lets `**` and `*` match segments that begin with `.`, which is
// required for `.env`, `.git`, etc.
const compile = (pattern: string) => picomatch(pattern, { dot: true });

const ALWAYS_PINNED = ["/rill.yaml"].map(compile);

const CLOUD_READONLY = ["**/.env", "**/.*.env"].map(compile);

const PROTECTED_DIRECTORIES = ["/tmp", "/tmp/**", "/.git", "/.git/**"].map(
  compile,
);

// Cloud injects this notice once at boot via setCloudReadonlyNotice. The
// snippet captures org/project from the layout scope, so its identity changes
// across layout instances; a non-reactive slot is sufficient because the
// notice is only read at render time, well after the cloud layout has set it.
let cloudReadonlyNotice: Snippet | undefined;

export function setCloudReadonlyNotice(notice: Snippet | undefined) {
  cloudReadonlyNotice = notice;
}

export function isManaged(path: string): boolean {
  if (isCloudRuntimeEditEnvironment()) {
    return CLOUD_READONLY.some((m) => m(path));
  }
  return false;
}

export function isPinned(path: string): boolean {
  return ALWAYS_PINNED.some((m) => m(path));
}

export function getReadonlyNotice(path: string): Snippet | undefined {
  if (isCloudRuntimeEditEnvironment() && CLOUD_READONLY.some((m) => m(path))) {
    return cloudReadonlyNotice;
  }
  return undefined;
}

export function isProtectedDirectory(path: string): boolean {
  return PROTECTED_DIRECTORIES.some((m) => m(path));
}
