#!/bin/sh
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Creating OAuth clients..."

hydra clients create \
    --endpoint https://oidc-provider.example.com:4445 \
    --id org1 \
    --secret org1-secret \
    --grant-types client_credentials \
    --response-types token,code \
    --scope org_admin \
    --skip-tls-verify

hydra clients create \
    --endpoint https://oidc-provider.example.com:4445 \
    --id "National Bank" \
    --secret bank-secret \
    --grant-types client_credentials \
    --response-types token,code \
    --scope org_admin \
    --skip-tls-verify

hydra clients create \
    --endpoint https://oidc-provider.example.com:4445 \
    --id test_org \
    --secret test-org-secret \
    --grant-types client_credentials \
    --response-types token,code \
    --scope org_admin \
    --skip-tls-verify

hydra clients create \
    --endpoint https://oidc-provider.example.com:4445 \
    --id test_bank \
    --secret test-bank-secret \
    --grant-types client_credentials \
    --response-types token,code \
    --scope org_admin \
    --skip-tls-verify

hydra clients create \
    --endpoint https://oidc-provider.example.com:4445 \
    --id issuer_oidc4vc \
    --secret issuer-oidc4vc-secret \
    --grant-types authorization_code \
    --response-types code \
    --scope openid,profile,address \
    --callbacks https://localhost:4455/oidc/redirect \
    --skip-tls-verify

echo "Finished creating OAuth clients"
