import * as child_process from 'child_process';
import * as vscode from 'vscode';
import { gaddisCmd } from "./platform";

export function activateFormat(context: vscode.ExtensionContext) {
    context.subscriptions.push(vscode.languages.registerDocumentFormattingEditProvider('gaddis', {
        async provideDocumentFormattingEdits(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
            if (document.languageId !== 'gaddis') {
                return [];
            }
            try {
                return await formatDocument(document);
            } catch (error) {
                vscode.window.showErrorMessage(`Formatting failed: ${error}`);
                return []; // Return an empty array to indicate no edits.
            }
        },
    }));
}


function formatDocument(document: vscode.TextDocument): Promise<vscode.TextEdit[]> {
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
