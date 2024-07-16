const core = require('@actions/core');
const exec = require('@actions/exec');

const util = require('util');
const fetch = require('node-fetch');
const mkdir = util.promisify(require('fs').mkdir);
const writeFile = util.promisify(require('fs').writeFile);
const path = require('path');


const MIRROR_DIR = 'network-mirror';


function semverKey(version) {
    const semver = version.split('-')[0];
    const [major, minor, patch] = semver.split('.');
    return parseInt(major) * 10000 + parseInt(minor) * 100 + parseInt(patch);
}


async function main() {
    core.startGroup('Building network mirror index');
    try {
        const registryDomain = core.getInput('registry-domain');
        const GCSBucket = core.getInput('gcs-bucket');
        const dryRun = core.getBooleanInput('dry-run');

        const response = await fetch(`https://${registryDomain}/scalr/scalr/versions`);
        const data = await response.json();

        const versions = data.versions.sort((a, b) => {
            const keyA = semverKey(a.version);
            const keyB = semverKey(b.version);
            return keyA < keyB ? -1 : keyA > keyB ? 1 : 0;
        });

        await mkdir(path.join(MIRROR_DIR, registryDomain, 'scalr', 'scalr'), { recursive: true });

        for (const { version, platforms } of versions) {
            console.log(`Processing ${version}`);
            const versionData = { archives: {} };

            for (const platform of platforms) {
                const { os: os_, arch } = platform;
                const platformName = `${os_}_${arch}`;
                versionData.archives[platformName] = {
                    url: `https://${registryDomain}/terraform-provider-scalr/${version}/terraform-provider-scalr_${version}_${platformName}.zip`
                };
            }

            const versionFilePath = path.join(MIRROR_DIR, registryDomain, 'scalr', 'scalr', `${version}.json`);
            await writeFile(versionFilePath, JSON.stringify(versionData, null, 4));
        }

        const indexData = { versions: {} };
        for (const version of versions) {
            indexData.versions[version.version] = {};
        }

        const indexFilePath = path.join(MIRROR_DIR, registryDomain, 'scalr', 'scalr', 'index.json');
        await writeFile(indexFilePath, JSON.stringify(indexData, null, 4));

        const bucketPath = GCSBucket + '/providers';
        if (!dryRun) {
            try {
                await exec.exec(
                    'gsutil -m -h "Cache-Control:private, max-age=0, no-transform"'
                    + ` rsync -d -r ${MIRROR_DIR}/ ${bucketPath}/`
                );
            } catch (err) {
                console.warn(`Failed to upload file: ${err.message}`)
            }
        }

        console.log('Mirror operation completed successfully.');
    } catch (err) {
        return core.setFailed(`Failed to update network mirror: ${err.message}.`)
    } finally {
        core.endGroup();
    }
}

main();
