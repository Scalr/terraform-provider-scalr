#!/usr/bin/env bash

# Required env variables:
# DOMAIN, GPG_KEY_ID, GPG_PUB_KEY, BUCKET_NAME

PROVIDER_NAME="terraform-provider-scalr"
PROVIDER_SOURCE="scalr/scalr"
URL="https://$DOMAIN"
PROTOCOLS="[\"5.0\"]"

VERSION=$(PAGER= git tag --points-at HEAD)
VERSION=${VERSION:1}
TMP_DIR=$(mktemp -d -t scalr-provider-$VERSION-XXXXXXXXXXX)
PROVIDER_BIN_PATH=$TMP_DIR/$PROVIDER_NAME/$VERSION
DOWNLOAD_DIR=$TMP_DIR/$PROVIDER_SOURCE/$VERSION/download/

gsutil -m rsync -R $BUCKET_NAME $TMP_DIR

mkdir -p $PROVIDER_BIN_PATH
cp dist/*.zip dist/*_SHA256SUMS* $PROVIDER_BIN_PATH

mkdir -p $DOWNLOAD_DIR

for zip_name in $(ls dist/*.zip)
do
    shasum=$(sha256sum $zip_name | head -c 64)
    zip_name=${zip_name#"dist/"}
    os_arch=${zip_name#"${PROVIDER_NAME}_${VERSION}_"}
    os_arch=${os_arch%".zip"}
    os_arch=(${os_arch//_/ })

    mkdir -p $DOWNLOAD_DIR/${os_arch[0]}

    cat << EOF > $DOWNLOAD_DIR/${os_arch[0]}/${os_arch[1]}
{
    "protocols": $PROTOCOLS,
    "os": "${os_arch[0]}",
    "arch": "${os_arch[1]}",
    "filename": "$zip_name",
    "download_url": "$URL/$PROVIDER_NAME/$VERSION/$zip_name",
    "shasums_url": "$URL/$PROVIDER_NAME/$VERSION/${PROVIDER_NAME}_${VERSION}_SHA256SUMS",
    "shasums_signature_url": "$URL/$PROVIDER_NAME/$VERSION/${PROVIDER_NAME}_${VERSION}_SHA256SUMS.sig",
    "shasum": "$shasum",
    "signing_keys": {
        "gpg_public_keys": [
            {
                "key_id": "$GPG_KEY_ID",
                "ascii_armor": "$GPG_PUB_KEY",
                "trust_signature": "",
                "source": "HashiCorp",
                "source_url": "https://www.hashicorp.com/security.html"
            }
        ]
    }
}
EOF
done

versions=()
for version in $(find "$TMP_DIR/$PROVIDER_SOURCE/" -maxdepth 1 ! -path "$TMP_DIR/$PROVIDER_SOURCE/" -type d -printf "%f\n")
do
    platforms=()
    for zip_name in $(ls dist/*.zip)
    do
        os_arch=${zip_name#"dist/${PROVIDER_NAME}_${VERSION}_"}
        os_arch=${os_arch%".zip"}
        os_arch=(${os_arch//_/ })
        platforms+=( "{\"os\": \"${os_arch[0]}\", \"arch\": \"${os_arch[1]}\"}" )
    done

    platforms=$(printf ", %s" "${platforms[@]}")
    platforms=${platforms:1}

    versions+=( "{\"version\": \"$version\", \"protocols\": $PROTOCOLS, \"platforms\": [$platforms]}" )
done

versions=$(printf ", %s" "${versions[@]}")
versions=${versions:1}

cat << EOF > $TMP_DIR/$PROVIDER_SOURCE/versions
{"versions": [$versions]}
EOF

mkdir -p $TMP_DIR/.well-known/

cat << EOF > $TMP_DIR/.well-known/terraform.json
{"providers.v1": "/"}
EOF

cat << EOF > $TMP_DIR/index.html
<p>Example usage:</p>
<pre>
<code>
terraform {
    required_providers {
        scalr = {
            source = "$URL/$PROVIDER_SOURCE"
            version= "$VERSION"
        }
    }
}
</code>
</pre>
EOF

gsutil -m -h "Cache-Control:private, max-age=0, no-transform" rsync -R $TMP_DIR $BUCKET_NAME

rm -rf $TMP_DIR
