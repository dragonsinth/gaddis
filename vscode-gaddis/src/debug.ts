import * as vscode from 'vscode';
import { CancellationToken, DebugConfiguration, ProviderResult, WorkspaceFolder } from 'vscode';

export function activateDebug(context: vscode.ExtensionContext) {
    const subs = context.subscriptions;

    subs.push(
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
                    program: targetResource.fsPath
                },
                    { noDebug: true }
                ).then(v => console.log(v));
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
                }).then(v => console.log(v));
            }
        }),
        vscode.debug.registerDebugConfigurationProvider('gaddis', new GaddisConfigurationProvider()),
    );
}


class GaddisConfigurationProvider implements vscode.DebugConfigurationProvider {

    /**
     * Massage a debug configuration just before a debug session is being launched,
     * e.g. add all missing attributes to the debug configuration.
     */
    resolveDebugConfiguration(folder: WorkspaceFolder | undefined, config: DebugConfiguration, token?: CancellationToken): ProviderResult<DebugConfiguration> {

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

        return config;
    }
}