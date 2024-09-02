export function getRepoNameFromGithubUrl(githubUrl: string) {
  const repoName = githubUrl.split("github.com/")[1];
  // remove trailing forwards slash if present
  return repoName?.replace(/\/$/, "") ?? "";
}
