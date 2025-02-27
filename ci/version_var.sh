#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#


# Release Parameters
BASE_VERSION=0.1.9
IS_RELEASE=false

# Project Parameters
BASE_VC_PKG_NAME=vc-server
BASE_COMPARATOR_PKG_NAME=comparator-server
BASE_CONFIDENTIAL_STORAGE_HUB_PKG_NAME=hub-confidential-storage
BASE_VAULT_PKG_NAME=vault-server
RELEASE_REPO=ghcr.io/trustbloc
SNAPSHOT_REPO=ghcr.io/trustbloc-cicd

if [ ${IS_RELEASE} = false ]
then
  EXTRA_VERSION=snapshot-$(git rev-parse --short=7 HEAD)
  PROJECT_VERSION=${BASE_VERSION}-${EXTRA_VERSION}
  PROJECT_PKG_REPO=${SNAPSHOT_REPO}
else
  PROJECT_VERSION=${BASE_VERSION}
  PROJECT_PKG_REPO=${RELEASE_REPO}
fi

export VC_SERVER_TAG=$PROJECT_VERSION
export DID_RESOLVER_TAG=$PROJECT_VERSION
export COMPARATOR_SERVER_TAG=$PROJECT_VERSION
export VAULT_SERVER_TAG=$PROJECT_VERSION
export VC_SERVER_PKG=${PROJECT_PKG_REPO}/${BASE_VC_PKG_NAME}
export COMPARATOR_SERVER_PKG=${PROJECT_PKG_REPO}/${BASE_COMPARATOR_PKG_NAME}
export VAULT_SERVER_PKG=${PROJECT_PKG_REPO}/${BASE_VAULT_PKG_NAME}
export CONFIDENTIAL_STORAGE_HUB_PKG=${PROJECT_PKG_REPO}/${BASE_CONFIDENTIAL_STORAGE_HUB_PKG_NAME}
export CONFIDENTIAL_STORAGE_HUB_TAG=$PROJECT_VERSION
