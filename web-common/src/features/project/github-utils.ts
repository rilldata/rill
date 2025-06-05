export function getRepoNameFromGitRemote(gitRemote: string) {
  let repoName = gitRemote.split("github.com/")[1];
  // remove trailing forwards slash if present
  repoName = repoName?.replace(/\/$/, "") ?? "";
  // remote .git suffix
  repoName = repoName?.replace(/\.git$/, "") ?? "";
  return repoName;
}
