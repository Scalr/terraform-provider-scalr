const core = require('@actions/core');
const exec = require('@actions/exec');

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { promisify } = require('util');

const mkdir = promisify(fs.mkdir);
const writeFile = promisify(fs.writeFile);
const readFile = promisify(fs.readFile);
const readdir = promisify(fs.readdir);
const copyFile = promisify(fs.copyFile);

const SOURCE = 'Scalr';
const SOURCE_URL = 'https://scalr.com';
const PROVIDER_NAME = 'terraform-provider-scalr';
const PROVIDER_SOURCE = 'scalr/scalr';
const PROTOCOLS = ['5.0'];

async function getVersion() {
    const distFiles = await readdir('dist');
    const sumsFile = distFiles.find((f) => f.endsWith('_SHA256SUMS'));
    if (!sumsFile) {
        throw new Error('SHA256SUMS file not found in dist/');
    }
    const match = sumsFile.match(/^terraform-provider-scalr_(.*)_SHA256SUMS$/);
    if (!match) {
        throw new Error('Could not parse version from SHA256SUMS filename');
    }
    return match[1];
}

async function sha256(filePath) {
    const content = await readFile(filePath);
    return crypto.createHash('sha256').update(content).digest('hex');
}

function parseOsArch(zipName, version) {
    const prefix = `${PROVIDER_NAME}_${version}_`;
    const suffix = '.zip';
    const osArch = zipName.replace(prefix, '').replace(suffix, '');
    const [os, arch] = osArch.split('_');
    return { os, arch };
}

async function main() {
    core.startGroup('Uploading provider to registry');
    try {
        const domain = core.getInput('registry-domain');
        const bucketName = core.getInput('gcs-bucket');
        const gpgKeyId = core.getInput('gpg-key-id');
        const gpgPubKey = core.getInput('gpg-pub-key');

        const url = `https://${domain}`;
        const version = await getVersion();
        const tmpDir = `scalr-provider-${version}-${Date.now()}`;

        console.log(`Starting to push ${PROVIDER_NAME}:${version}`);

        // Copy remote provider registry to working directory
        // Old versions of terraform provider are used for composing versions file
        await mkdir(tmpDir, { recursive: true });
        await exec.exec(`gcloud storage rsync --exclude ".*zip$" --recursive ${bucketName} ${tmpDir}`);

        // Copy all built binaries, sha256sums and its signature to bin path
        const providerBinPath = path.join(tmpDir, PROVIDER_NAME, version);
        await mkdir(providerBinPath, { recursive: true });

        const distFiles = await readdir('dist');
        const filesToCopy = distFiles.filter((f) => f.endsWith('.zip') || f.includes('_SHA256SUMS'));
        for (const file of filesToCopy) {
            await copyFile(path.join('dist', file), path.join(providerBinPath, file));
        }

        // Create download directory for provider package metadata
        const downloadDir = path.join(tmpDir, PROVIDER_SOURCE, version, 'download');
        await mkdir(downloadDir, { recursive: true });

        // Compose provider package metadata for each arch and os
        // https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
        const zipFiles = distFiles.filter((f) => f.endsWith('.zip'));
        const platforms = [];

        for (const zipName of zipFiles) {
            const { os, arch } = parseOsArch(zipName, version);
            const shasum = await sha256(path.join('dist', zipName));

            platforms.push({ os, arch });

            const packageMetadata = {
                protocols: PROTOCOLS,
                os,
                arch,
                filename: zipName,
                download_url: `${url}/${PROVIDER_NAME}/${version}/${zipName}`,
                shasums_url: `${url}/${PROVIDER_NAME}/${version}/${PROVIDER_NAME}_${version}_SHA256SUMS`,
                shasums_signature_url: `${url}/${PROVIDER_NAME}/${version}/${PROVIDER_NAME}_${version}_SHA256SUMS.sig`,
                shasum,
                signing_keys: {
                    gpg_public_keys: [
                        {
                            key_id: gpgKeyId,
                            ascii_armor: gpgPubKey,
                            trust_signature: '',
                            source: SOURCE,
                            source_url: SOURCE_URL,
                        },
                    ],
                },
            };

            const osDir = path.join(downloadDir, os);
            await mkdir(osDir, { recursive: true });
            await writeFile(path.join(osDir, arch), JSON.stringify(packageMetadata, null, 4));
        }

        // Compose file with all available terraform provider versions and supported platforms
        // https://www.terraform.io/docs/internals/provider-registry-protocol.html#list-available-versions
        const providerSourceDir = path.join(tmpDir, PROVIDER_SOURCE);
        const versionDirs = (await readdir(providerSourceDir, { withFileTypes: true }))
            .filter((dirent) => dirent.isDirectory())
            .map((dirent) => dirent.name);

        const versions = versionDirs.map((ver) => ({
            version: ver,
            protocols: PROTOCOLS,
            platforms,
        }));

        await writeFile(
            path.join(providerSourceDir, 'versions'),
            JSON.stringify({ versions }, null, 4)
        );

        // Create .well-known/terraform.json
        const wellKnownDir = path.join(tmpDir, '.well-known');
        await mkdir(wellKnownDir, { recursive: true });
        await writeFile(
            path.join(wellKnownDir, 'terraform.json'),
            JSON.stringify({ 'providers.v1': '/' }, null, 4)
        );

        // Compose index.html page with provider example usage
        const providerDir = path.join(tmpDir, PROVIDER_NAME);
        const availableVersions = (await readdir(providerDir, { withFileTypes: true }))
            .filter((dirent) => dirent.isDirectory())
            .map((dirent) => dirent.name);

        const indexHtml = `<html>
    <head>
        <meta charset='UTF-8'>
        <title>Scalr terraform registry</title>
    </head>
    <body>
        <p>Versions:</p>
        <p><code>${availableVersions.join(', ')}</code></p>
        <p>Example:</p>
        <pre>
            <code>
terraform {
    required_providers {
        scalr = {
            source = "${domain}/${PROVIDER_SOURCE}"
            version= "${version}"
        }
    }
}
            </code>
        </pre>
    </body>
</html>
`;

        await writeFile(path.join(tmpDir, 'index.html'), indexHtml);

        // Upload to GCS with no caching
        await exec.exec(
            `gcloud storage rsync --cache-control="private, max-age=0, no-transform" --recursive ${tmpDir} ${bucketName}`
        );

        // Clean up
        await fs.promises.rm(tmpDir, { recursive: true, force: true });

        console.log('Provider upload completed successfully.');
    } catch (err) {
        return core.setFailed(`Failed to upload provider: ${err.message}`);
    } finally {
        core.endGroup();
    }
}

main();
