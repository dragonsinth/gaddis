import * as os from 'os';
import * as child_process from 'child_process';
import * as path from 'path';
import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
    const window = vscode.window
    const subs = context.subscriptions;
    const gaddisCmd = getGaddisExecutablePath(context);

    vscode.languages.setLanguageConfiguration('gaddis', {
        indentationRules: {
            increaseIndentPattern: /^\s*(While|Do|For|Select|Case|Default|If|Else|Module|Function|Class).*$/,
            decreaseIndentPattern: /^\s*((End|Until|Else|Case)\b)/
        }
    });

    const format = vscode.languages.registerDocumentFormattingEditProvider('gaddis', {
        async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
            try {
                return await formatDocument(gaddisCmd, document);
            } catch (error) {
                window.showErrorMessage(`Formatting failed: ${error}`);
                return []; // Return an empty array to indicate no edits.
            }
        },
    });
    subs.push(format);

    const diagnosticCollection = vscode.languages.createDiagnosticCollection('gaddis');

    const dirty: Record<string, vscode.TextDocument> = {}

    async function runDiagnostics(document: vscode.TextDocument, reason: string) {
        try {
            if (document.isClosed) {
                return
            }
            const diagnostics = await runCheck(gaddisCmd, document);
            diagnosticCollection.set(document.uri, diagnostics);
        } catch (error) {
            window.showErrorMessage(`Check failed: ${error}`);
            diagnosticCollection.clear();
        } finally {
            delete dirty[document.fileName];
        }
    }

    let flushTimeout: NodeJS.Timeout | null = null;

    function flushDiagnostics() {
        const todo = Object.entries(dirty)
        flushTimeout = null;
        for (const entry of todo) {
            runDiagnostics(entry[1], 'flush').catch(() => {
            });
        }
    }

    // Run diagnostics on document open, save, and change
    vscode.workspace.onDidOpenTextDocument(d => runDiagnostics(d, 'open'));
    vscode.workspace.onDidSaveTextDocument(d => runDiagnostics(d, 'save'));
    vscode.workspace.onDidChangeTextDocument((event) => {
        const doc = event.document;
        dirty[doc.fileName] = doc;
        if (flushTimeout != null) {
            clearTimeout(flushTimeout);
        }
        flushTimeout = setTimeout(flushDiagnostics, 1000)
    });

    // Run diagnostics on activation for the active document.
    if (window.activeTextEditor) {
        runDiagnostics(window.activeTextEditor.document, 'activate');
    }
    subs.push(diagnosticCollection);
}

function formatDocument(gaddisCmd: string, document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
    return new Promise((resolve, reject) => {
        const process = child_process.spawn(gaddisCmd, ['format']);
        let formattedText = '';
        let errorOutput = '';

        process.stdout.on('data', (data) => {
            formattedText += data.toString();
        });

        process.stderr.on('data', (data) => {
            errorOutput += data.toString();
        });

        process.on('close', (code) => {
            if (code === 0) {
                const fullRange = new vscode.Range(
                    document.positionAt(0),
                    document.positionAt(document.getText().length)
                );
                resolve([vscode.TextEdit.replace(fullRange, formattedText)]);
            } else {
                reject(errorOutput);
            }
        });

        process.on('error', (err) => {
            reject(err.message);
        });

        process.stdin.write(document.getText());
        process.stdin.end();
    });
}

async function runCheck(gaddisCmd: string, document: vscode.TextDocument): Promise<vscode.Diagnostic[]> {
    return new Promise((resolve, reject) => {
        const process = child_process.spawn(gaddisCmd, ['-json', 'check']);
        let output = '';
        let errorOutput = '';

        process.stdout.on('data', (data) => {
            output += data.toString();
        });

        process.stderr.on('data', (data) => {
            errorOutput += data.toString();
        });

        process.on('close', (code) => {
            if (code === 0) {
                try {
                    const diagnostics: vscode.Diagnostic[] = JSON.parse(output);
                    resolve(diagnostics);
                } catch (error) {
                    reject('Error parsing JSON: ' + error);
                }
            } else {
                reject(errorOutput);
            }
        });

        process.on('error', (err) => {
            reject(err.message);
        });

        process.stdin.write(document.getText());
        process.stdin.end();
    });
}


const OS_MAP: { [key: string]: string } = {
    win32: 'windows',
    darwin: 'darwin',
    linux: 'linux',
};

const ARCH_MAP: { [key: string]: string } = {
    arm64: 'arm64',
    x64: 'amd64',
};

function getGaddisExecutablePath(context: vscode.ExtensionContext): string {
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
