import type { Snippet } from "svelte";
import { getRuntimeEditEnvironment } from "../edit-environment.ts";

// Two distinct kinds of path protections:
//
// - readonly: file content is locked (editor disables typing). Renames and
//   deletes are also blocked because you can't act on a file you can't edit.
//   `.env` on cloud is the canonical case.
//
// - pinned: content is editable, but the path is locked. The file can't be
//   renamed or deleted, and other files can't be renamed onto it. `rill.yaml`
//   is the canonical case.
//
// The two are surfaced as separate predicates so that call sites express what
// they actually need, rather than a generic "protected" flag with hidden flips.

const ALWAYS_PINNED: string[] = ["/rill.yaml"];

export const CLOUD_READONLY: string[] = ["**/.env", "**/.*.env"];

const PROTECTED_DIRECTORIES: string[] = [
  "/tmp",
  "/tmp/**",
  "/.git",
  "/.git/**",
];

// Cloud injects this notice once at boot via setCloudReadonlyNotice. The
// snippet captures org/project from the layout scope, so its identity changes
// across layout instances; a non-reactive slot is sufficient because the
// notice is only read at render time, well after the cloud layout has set it.
let cloudReadonlyNotice: Snippet | undefined;

export function setCloudReadonlyNotice(notice: Snippet | undefined) {
  cloudReadonlyNotice = notice;
}

export function isReadonly(path: string): boolean {
  if (getRuntimeEditEnvironment() === "cloud") {
    if (CLOUD_READONLY.some((p) => matchGlob(p, path))) return true;
  }
  return false;
}

export function isPinned(path: string): boolean {
  return ALWAYS_PINNED.some((p) => matchGlob(p, path));
}

export function getReadonlyNotice(path: string): Snippet | undefined {
  if (getRuntimeEditEnvironment() === "cloud") {
    if (CLOUD_READONLY.some((p) => matchGlob(p, path))) {
      return cloudReadonlyNotice;
    }
  }
  return undefined;
}

export function isProtectedDirectory(path: string): boolean {
  return PROTECTED_DIRECTORIES.some((p) => matchGlob(p, path));
}

// Minimal glob → RegExp compiler. Supports the small vocabulary we need:
//   `**` matches any sequence of characters, including `/`.
//   `*`  matches any sequence not crossing `/`.
//   `?`  matches a single character not equal to `/`.
// All other characters are matched literally. Patterns are anchored.
const compiledGlobs = new Map<string, RegExp>();

function matchGlob(pattern: string, path: string): boolean {
  let regex = compiledGlobs.get(pattern);
  if (!regex) {
    regex = compileGlob(pattern);
    compiledGlobs.set(pattern, regex);
  }
  return regex.test(path);
}

function compileGlob(pattern: string): RegExp {
  let out = "^";
  for (let i = 0; i < pattern.length; i++) {
    const c = pattern[i];
    if (c === "*") {
      if (pattern[i + 1] === "*") {
        out += ".*";
        i++;
      } else {
        out += "[^/]*";
      }
    } else if (c === "?") {
      out += "[^/]";
    } else if (/[.+^${}()|[\]\\]/.test(c)) {
      out += "\\" + c;
    } else {
      out += c;
    }
  }
  return new RegExp(out + "$");
}
