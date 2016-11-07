#!/usr/bin/env sh

set -e -u

GH_API_BASE="https://api.github.com/repos/hellofresh/kandalf"

# Fetch the list of github pre-releases
PRE_RELEASES=$(\
    curl -s "$GH_API_BASE/releases?access_token=$GITHUB_TOKEN" \
    | jq "map(select(.prerelease==true)) | .[$MAX_RELEASES:] | .[] | .id" \
)

# Delete old releases
for release in ${PRE_RELEASES}
do
    echo "Deleting release #$release"
    curl -s -XDELETE "$GH_API_BASE/releases/$release?access_token=$GITHUB_TOKEN"
done
