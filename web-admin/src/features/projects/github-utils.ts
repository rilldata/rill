export function getRepoNameFromGithubUrl(githubUrl: string) {
  const repoName = githubUrl.split("github.com/")[1];
  return repoName;
}
