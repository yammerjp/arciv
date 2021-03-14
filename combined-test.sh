#!/bin/bash -e

# The shell script is the combined-test of arciv
# Unit tests is implemented by golang

if ! command -v sha256sum; then
  echo "Need 'sha256sum'. excute '\$brew install coreutils' if your computer is macOS"
  exit 1
fi

# Build
project_dir="$(cd $(dirname "$0"); pwd)"
cd "${project_dir}"

go build
arciv_bin="${project_dir}/arciv"

# Initialize directories
testing_tmp_dir="/tmp/arciv-combined-test"
local_repo_dir="${testing_tmp_dir}/local-repo"
remote_repo_dir="${testing_tmp_dir}/remote-repo"
rm -rf /tmp/arciv-combined-test
mkdir -p ${local_repo_dir}
cd ${local_repo_dir}

# Add files to the local repository
cp -r "${project_dir}/commands" "${local_repo_dir}"

# Snapshot
find "${local_repo_dir}" -type f -print0 | xargs -0 sha256sum > "${testing_tmp_dir}/beforeStore"

# Initialize a local repository
${arciv_bin} init

${arciv_bin} repository add "name:remote-repo" "path:${remote_repo_dir}" "type:file"

# Store blobs to a remote repository
${arciv_bin} store --repository remote-repo

# Delete files of the local repository
mv "${local_repo_dir}/.arciv" "${testing_tmp_dir}/.arciv"
rm -rf "${local_repo_dir}"
mkdir "${local_repo_dir}"
mv "${testing_tmp_dir}/.arciv" "${local_repo_dir}/.arciv"

cd ${local_repo_dir}
# Restore blobs from the remote repository to the local repository 
${arciv_bin} restore --repository remote-repo --commit "$(${arciv_bin} log --repository remote-repo)" -f

# Remove .arciv
rm -rf "${local_repo_dir}/.arciv"

# Snapshot
find "${local_repo_dir}" -type f -print0 | xargs -0 sha256sum > "${testing_tmp_dir}/afterRestore"

# Compare shapshots
if diff "${testing_tmp_dir}/beforeStore" "${testing_tmp_dir}/afterRestore"; then
  echo "success"
else 
  echo "failed"
fi
