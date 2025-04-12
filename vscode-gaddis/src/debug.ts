import * as vscode from 'vscode';
import * as path from 'path';
import { dapServerPort } from "./dapServer";

export function activateDebug(context: vscode.ExtensionContext) {
    context.subscriptions.push(
        vscode.debug.registerDebugConfigurationProvider('gaddis', {
            resolveDebugConfiguration(folder: vscode.WorkspaceFolder | undefined, config: vscode.DebugConfiguration, token?: vscode.CancellationToken): vscode.ProviderResult<vscode.DebugConfiguration> {
                // if launch.json is missing or empty
                if (!config.type && !config.request && !config.name) {
                    const editor = vscode.window.activeTextEditor;
                    if (editor && editor.document.languageId === 'gaddis') {
                        config.type = 'gaddis';
                        config.name = 'Launch';
                        config.request = 'launch';
                        config.program = '${file}';
                        config.stopOnEntry = true;
                        config.testMode = false;
                    }
                }

                if (!config.program) {
                    return vscode.window.showInformationMessage("Cannot find a program to debug").then(_ => {
                        return undefined;	// abort launch
                    });
                }
                if (!config.workDir) {
                    config.workDir = '${workspaceFolder}';
                }
                if (!config.debugServer) {
                    if (!dapServerPort) {
                        return vscode.window.showWarningMessage("Debug server not running").then(_ => {
                            return undefined;	// abort launch
                        });
                    }
                    config.debugServer = dapServerPort;
                }
                return config;
            },
        }),
        vscode.commands.registerCommand('extension.gaddis.runEditorContents', (resource: vscode.Uri) => {
            let targetResource = resource;
            if (!targetResource && vscode.window.activeTextEditor) {
                targetResource = vscode.window.activeTextEditor.document.uri;
            }
            if (targetResource) {
                vscode.debug.startDebugging(undefined, {
                    type: 'gaddis',
                    name: 'Run File',
                    request: 'launch',
                    program: targetResource.fsPath,
                    workDir: path.dirname(targetResource.fsPath),
                },
                    { noDebug: true }
                );
            }
        }),
        vscode.commands.registerCommand('extension.gaddis.debugEditorContents', (resource: vscode.Uri) => {
            let targetResource = resource;
            if (!targetResource && vscode.window.activeTextEditor) {
                targetResource = vscode.window.activeTextEditor.document.uri;
            }
            if (targetResource) {
                vscode.debug.startDebugging(undefined, {
                    type: 'gaddis',
                    name: 'Debug File',
                    request: 'launch',
                    program: targetResource.fsPath,
                    workDir: path.dirname(targetResource.fsPath),
                    stopOnEntry: false,
                });
            }
        }),
        vscode.commands.registerCommand('extension.gaddis.testEditorContents', (resource: vscode.Uri) => {
            let targetResource = resource;
            if (!targetResource && vscode.window.activeTextEditor) {
                targetResource = vscode.window.activeTextEditor.document.uri;
            }
            if (targetResource) {
                vscode.debug.startDebugging(undefined, {
                    type: 'gaddis',
                    name: 'Test File',
                    request: 'launch',
                    program: targetResource.fsPath,
                    workDir: path.dirname(targetResource.fsPath),
                    testMode: true,
                },
                    { noDebug: true }
                );
            }
        }),
    );
}
