PATCH_TAG="v0.60.6"
RELEASE_BRANCH="release-0.60" 

git checkout ${RELEASE_BRANCH}
git tag -a ${PATCH_TAG} -m "${PATCH_TAG} release"
git push origin ${PATCH_TAG}
