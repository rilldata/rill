#!/usr/bin/env bash
set -e

# Create Metadata
if [[ -z "${TRAVIS_TAG}" ]]; then
  export VERSION=nightly
else
  export VERSION=${TRAVIS_TAG}
fi

echo "version=${VERSION}" > metadata.txt;
echo "build_url=${TRAVIS_BUILD_WEB_URL}" >> metadata.txt
echo "sha=$(git rev-parse HEAD)" >> metadata.txt
echo "time=$(date +%FT%T)" >> metadata.txt

# Activate GCP Access
echo ${GCP_TOKEN} > gcp.json
gcloud auth activate-service-account --key-file gcp.json
gcloud config set project rilldata

# Upload binary
upload(){
  file=$1
  path=$2

  echo "Uploading ${file} to gs://prod-cdn.rilldata.com/rill/${VERSION}/${path}"
  gsutil cp ${file} gs://prod-cdn.rilldata.com/rill/${VERSION}/${path}

  if [[ "${VERSION}" != "nightly" ]]; then
    echo "Uploading ${file} to gs://prod-cdn.rilldata.com/rill/latest/${path}"
    gsutil cp ${file} gs://prod-cdn.rilldata.com/rill/latest/${path}
  fi
}

if [[ ${TRAVIS_OS_NAME} == "osx" ]]; then
  shasum -a 256 rilldata/rill-macos-x64 > rill.sha256
  upload rilldata/rill-macos-x64 macos-x64/rill
  upload rill.sha256 macos-x64/rill.sha256
  upload metadata.txt metadata.txt
fi

if [[ ${TRAVIS_OS_NAME} == "linux" ]]; then
  sha256sum rilldata/rill-linux-x64 > rill.sha256
  upload rilldata/rill-linux-x64 linux-x64/rill
  upload rill.sha256 linux-x64/rill.sha256
fi

if [[ ${TRAVIS_OS_NAME} == "windows" ]]; then
  cp gcp.json /c/gcp.json
  echo -e "[Credentials]\ngs_service_key_file=c:/gcp.json" > /c/.boto
  export BOTO_CONFIG="c:/.boto"

  gsutil() {
    /c/Program\ Files\ \(x86\)/Google/Cloud\ SDK/google-cloud-sdk/platform/gsutil_py2/gsutil $1 $2 $3
  }

  CertUtil -hashfile rilldata/rill-win-x64.exe SHA256 > rill.sha256

  upload rilldata/rill-win-x64.exe win-x64/rill.exe
  upload rill.sha256 win-x64/rill.sha256
fi
