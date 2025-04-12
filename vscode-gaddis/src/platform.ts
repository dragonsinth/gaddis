import os from "os";
import path from "path";

const OS_MAP: { [key: string]: string } = {
    win32: 'windows',
    darwin: 'darwin',
    linux: 'linux',
};

const ARCH_MAP: { [key: string]: string } = {
    arm64: 'arm64',
    x64: 'amd64',
};

function getGaddisExecutablePath(): string {
    const platform = os.platform();
    const arch = os.arch();

    const goOS = OS_MAP[platform];
    if (!goOS) {
        throw new Error(`Unsupported platform: ${platform}`);
    }
    const goArch = ARCH_MAP[arch];
    if (!goArch) {
        throw new Error(`Unsupported architecture: ${arch}`);
    }
    const ext = (goOS == 'windows') ? '.exe' : ''
    return path.join(__dirname, '..', 'bin', `gaddis-${goOS}-${goArch}${ext}`);
}

export const gaddisCmd = getGaddisExecutablePath();
