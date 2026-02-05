const githubUrlPrefixRegex = /github\.com\/|git@github\.com:/;
const githubUrlGitSuffixRegex = /\.git$/;

export function getRepoNameFromGitRemote(gitRemote: string) {
  let repoName = gitRemote.split(githubUrlPrefixRegex)[1];
  // remove trailing forwards slash if present
  repoName = repoName?.replace(/\/$/, "") ?? "";
  // remote .git suffix
  repoName = repoName?.replace(githubUrlGitSuffixRegex, "") ?? "";
  return repoName;
}

export function getGitUrlFromRemote(remote: string | undefined) {
  return remote?.replace(githubUrlGitSuffixRegex, "");
}

const MergeConflictsError =
  /Your local changes to the following files would be overwritten by merge/;

export function isMergeConflictError(errorMessage: string) {
  return MergeConflictsError.test(errorMessage);
}
