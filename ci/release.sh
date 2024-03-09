#!/usr/bin/env bash

gh api \
  --method POST \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  /repos/SiasMey/notebox/releases \
  -f tag_name='v1.4.1' \
 -f target_commitish='trunk' \
 -f name='v1.4.1' \
 -f body='Description of the release' \
 -F draft=false \
 -F prerelease=false \
 -F generate_release_notes=false 
