import * as child_process from 'child_process';
import * as vscode from 'vscode';
import {getGaddisExecutablePath} from './platform';
import {GaddisRunProvider} from "./run_provider";

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

    subs.push(vscode.languages.registerDocumentFormattingEditProvider('gaddis', {
        async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
            if (document.languageId !== 'gaddis') {
                return [];
            }
            try {
                return await formatDocument(gaddisCmd, document);
            } catch (error) {
                window.showErrorMessage(`Formatting failed: ${error}`);
                return []; // Return an empty array to indicate no edits.
            }
        },
    }));

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
            runDiagnostics(entry[1]).catch(() => {
            });
        }
    }

    // Run diagnostics on document open, save, and change
    vscode.workspace.onDidOpenTextDocument(d => runDiagnostics(d));
    vscode.workspace.onDidSaveTextDocument(d => runDiagnostics(d));
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
        runDiagnostics(window.activeTextEditor.document);
    }
    subs.push(diagnosticCollection);

    const gaddisRunProvider = new GaddisRunProvider(gaddisCmd);
    subs.push(vscode.tasks.registerTaskProvider('gaddis', gaddisRunProvider));

    subs.push(vscode.commands.registerCommand('gaddis.runTask', () => {
      vscode.tasks.fetchTasks({ type: 'gaddis' }).then((tasks) => {
        if (tasks && tasks.length > 0) {
          vscode.tasks.executeTask(tasks[0]);
        }
      });
    }));

    subs.push(vscode.commands.registerCommand('myExtension.runGadTask', (fileUri: vscode.Uri) => {
      // Logic to run your task on the specified .gad file
      // You can use fileUri.fsPath to get the file's path
      vscode.tasks.fetchTasks({ type: 'myGadTask' }).then((tasks) => {
        if (tasks && tasks.length > 0) {
          //If your task requires the file path, modify the task appropriately here.
          //For example, add a command line argument.
          vscode.tasks.executeTask(tasks[0]);
        } else {
          vscode.window.showErrorMessage("Gad task not found");
        }
      });
    }));
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
