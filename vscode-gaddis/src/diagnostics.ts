import * as child_process from 'child_process';
import * as vscode from 'vscode';
import { gaddisCmd } from "./platform";

export function activateDiagnostics(context: vscode.ExtensionContext) {
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('gaddis');

    const dirty: Record<string, vscode.TextDocument> = {}

    async function runDiagnostics(document: vscode.TextDocument) {
        if (document.languageId !== 'gaddis') {
            return [];
        }
        try {
            if (document.isClosed) {
                return
            }
            const diagnostics = await runCheck(document);
            diagnosticCollection.set(document.uri, diagnostics);
        } catch (error) {
            vscode.window.showErrorMessage(`Check failed: ${error}`);
            diagnosticCollection.clear();
        } finally {
            delete dirty[document.fileName];
        }
    }

    let flushTimeout: NodeJS.Timeout | null = null;

    async function flushDiagnostics() {
        const todo = Object.entries(dirty)
        flushTimeout = null;
        for (const entry of todo) {
            await runDiagnostics(entry[1]);
        }
    }

    // Run diagnostics on document open, save, and change
    context.subscriptions.push(
        diagnosticCollection,
        vscode.workspace.onDidOpenTextDocument(d => runDiagnostics(d)),
        vscode.workspace.onDidSaveTextDocument(d => runDiagnostics(d)),
        vscode.workspace.onDidChangeTextDocument((event) => {
            const doc = event.document;
            dirty[doc.fileName] = doc;
            if (flushTimeout != null) {
                clearTimeout(flushTimeout);
            }
            flushTimeout = setTimeout(flushDiagnostics, 1000)
        }),
    );

    // Run diagnostics on activation for the active document.
    if (vscode.window.activeTextEditor) {
        runDiagnostics(vscode.window.activeTextEditor.document);
    }
}

function runCheck(document: vscode.TextDocument): Promise<vscode.Diagnostic[]> {
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
