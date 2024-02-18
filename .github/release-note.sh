#!/usr/bin/env bash

RELEASE=${RELEASE:-$2}
PREVIOUS_RELEASE=${PREVIOUS_RELEASE:-$1}

# ref https://stackoverflow.com/questions/1441010/the-shortest-possible-output-from-git-log-containing-author-and-date
CHANGELOG=$(git log --no-merges --date=short --pretty=format:'- %h %an %ad %s' "${PREVIOUS_RELEASE}".."${RELEASE}")

cat <<EOF
# SSHVPN release ${RELEASE}

SSHVPN ${RELEASE} is available now ! 🎉

## Download SSHVPN for your platform

**Mac** (x86-64/Intel)

\`\`\`
curl -Lo sshvpn.zip https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_darwin_amd64.zip && unzip -d sshvpn sshvpn.zip
\`\`\`

**Mac** (AArch64/Apple M1 silicon)

\`\`\`
curl -Lo sshvpn.zip https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_darwin_arm64.zip && unzip -d sshvpn sshvpn.zip
\`\`\`

**Linux** (x86-64)

\`\`\`
curl -Lo sshvpn.zip https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_linux_amd64.zip && unzip -d sshvpn sshvpn.zip
\`\`\`

**Linux** (AArch64)

\`\`\`
curl -Lo sshvpn.zip https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_linux_arm64.zip && unzip -d sshvpn sshvpn.zip
\`\`\`

**Linux** (i386)

\`\`\`
curl -Lo sshvpn.zip https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_linux_386.zip && unzip -d sshvpn sshvpn.zip
\`\`\`

**Windows** (x86-64)

\`\`\`
curl -LO https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_windows_amd64.zip
\`\`\`

**Windows** (AArch64)

\`\`\`
curl -LO https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_windows_arm64.zip
\`\`\`

**Windows** (i386)

\`\`\`
curl -LO https://github.com/wencaiwulue/sshvpn/releases/download/${RELEASE}/sshvpn_${RELEASE}_windows_386.zip
\`\`\`

## Checksums

SHA256 checksums available for compiled binaries.
Run \`shasum -a 256 -c checksums.txt\` to verify.

## Upgrading

Run \`sshvpn upgrade\` to upgrade from a previous version.

## Changelog

${CHANGELOG}
EOF
