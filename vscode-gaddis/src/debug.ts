import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';
import { gaddisCmd } from './platform';

let serverPort: number = 0;

export function activateDebug(context: vscode.ExtensionContext) {
    const outputChannel = vscode.window.createOutputChannel('Gaddis Debug Server', { log: true });
    // outputChannel.show();

    context.subscriptions.push(
        outputChannel,
        runServer(outputChannel),
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
                    if (!serverPort) {
                        return vscode.window.showWarningMessage("Debug server not running").then(_ => {
                            return undefined;	// abort launch
                        });
                    }
                    config.debugServer = serverPort;
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
                    stopOnEntry: true,
                });
            }
        }),
    );
}

function runServer(ch: vscode.LogOutputChannel): { dispose: () => void } {
    let serverProcess: child_process.ChildProcessWithoutNullStreams | null = null;
    let shouldStop = false;

    function doRunServer(ch: vscode.LogOutputChannel) {
        serverProcess = child_process.spawn(gaddisCmd, ["-port", "0", "debug"], { stdio: 'pipe' });

        serverProcess.stdout.on('data', streamToLineAdapter(line => {
            ch.trace(line);
        }));

        serverProcess.stderr.on('data', streamToLineAdapter(line => {
            if (serverPort == 0) {
                serverPort = parsePort(line);
            }
            ch.info(line);
        }));

        serverProcess.on('exit', (code: any) => {
            ch.appendLine(`Debug server exited with code ${code}; restarting.`);
            serverProcess = null;
            serverPort = 0;
            if (!shouldStop) {
                doRunServer(ch);
            }
        });
    }

    doRunServer(ch);
    return {
        dispose: () => {
            shouldStop = true;
            if (serverProcess) {
                serverProcess.kill('SIGTERM');
            }
        },
    }
}

function parsePort(output: string): number {
    const match = output.match(/127\.0\.0\.1:(\d+)/);
    if (match && match[1]) {
        return parseInt(match[1], 10);
    }
    return 0;
}

function streamToLineAdapter(f: (line: string) => void): (data: any) => void {
    let buffer = '';
    return (data: any) => {
        buffer += data.toString();
        for (let pos = buffer.indexOf('\n'); pos >= 0; pos = buffer.indexOf('\n')) {
            const line = buffer.substring(0, pos).trim();
            buffer = buffer.substring(pos + 1)
            if (line) {
                f(line)
            }
        }
    }
}
