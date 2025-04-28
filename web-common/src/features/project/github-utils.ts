export function getRepoNameFromGithubUrl(githubUrl: string) {
  const repoName = githubUrl.split("github.com/")[1];
  // remove trailing forwards slash if present
  return repoName?.replace(/\/$/, "") ?? "";
}

export function isRillManagedGithubOrg(githubUrl: string) {
  try {
    const url = new URL(githubUrl);

    if (url.hostname !== "github.com") {
      return false;
    }

    const parts = url.pathname.split("/").filter(Boolean);
    if (parts.length !== 2) {
      return false;
    }

    const [account] = parts;
    return account.startsWith("anshul-test-pp");
  } catch (err) {
    return false;
  }
}
