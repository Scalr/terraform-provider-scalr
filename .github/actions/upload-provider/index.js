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
const DIST_DIR = 'dist';

async function getVersion() {
    const distFiles = await readdir(DIST_DIR);
    const sumsFile = distFiles.find(f => f.endsWith('_SHA256SUMS'));
    if (!sumsFile) {
        throw new Error(`SHA256SUMS file not found in ${DIST_DIR}/`);
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
    const osArch = zipName.replace(`${PROVIDER_NAME}_${version}_`, '').replace('.zip', '');
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

        const distFiles = await readdir(DIST_DIR);
        const filesToCopy = distFiles.filter(f => f.endsWith('.zip') || f.includes('_SHA256SUMS'));
        for (const file of filesToCopy) {
            await copyFile(path.join(DIST_DIR, file), path.join(providerBinPath, file));
        }

        // Create download directory for provider package metadata
        const downloadDir = path.join(tmpDir, PROVIDER_SOURCE, version, 'download');
        await mkdir(downloadDir, { recursive: true });

        // Compose provider package metadata for each arch and os
        // https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
        const zipFiles = distFiles.filter(f => f.endsWith('.zip'));
        const platforms = [];

        for (const zipName of zipFiles) {
            const { os, arch } = parseOsArch(zipName, version);
            const shasum = await sha256(path.join(DIST_DIR, zipName));

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
        const versions = (await readdir(providerSourceDir, { withFileTypes: true }))
            .filter(dirent => dirent.isDirectory())
            .map(dirent => dirent.name);

        const versionsFile = versions.map(ver => ({
            version: ver,
            protocols: PROTOCOLS,
            platforms,
        }));

        await writeFile(
            path.join(providerSourceDir, 'versions'),
            JSON.stringify({ versions: versionsFile }, null, 4)
        );

        // Create .well-known/terraform.json
        const wellKnownDir = path.join(tmpDir, '.well-known');
        await mkdir(wellKnownDir, { recursive: true });
        await writeFile(
            path.join(wellKnownDir, 'terraform.json'),
            JSON.stringify({ 'providers.v1': '/' }, null, 4)
        );

        const indexTemplate = await readFile(path.join(__dirname, 'index.html'), 'utf8');
        const versionTags = versions
            .map(v => `<span class="version-tag${v === version ? ' selected' : ''}" data-version="${v}">${v}</span>`)
            .join('');
        const indexHtml = indexTemplate
            .replace('{{VERSIONS}}', versionTags)
            .replace('{{VERSIONS}}', `<span class="version-tag selected" data-version="1.0.0">1.0.0</span>`)
            .replace('{{DOMAIN}}', domain)
            .replace('{{PROVIDER_SOURCE}}', PROVIDER_SOURCE)
            .replace('{{VERSION}}', version);

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
