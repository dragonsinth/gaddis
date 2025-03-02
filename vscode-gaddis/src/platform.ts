import vscode from "vscode";
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

export function getGaddisExecutablePath(context: vscode.ExtensionContext): string {
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
    return path.join(context.extensionPath, 'bin', `gaddis-${goOS}-${goArch}`);
}
