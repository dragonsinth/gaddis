import * as vscode from 'vscode';
import { CancellationToken, DebugConfiguration, ProviderResult, WorkspaceFolder } from 'vscode';
import { GaddisRunProvider } from "./run_provider";
import { GaddisTestProvider } from "./test_provider";
import { makeTask } from './task';
import { activateDiagnostics } from './diagnostics';
import { activateFormat } from './format';

export function activate(context: vscode.ExtensionContext) {
    const subs = context.subscriptions;

    subs.push(
        vscode.languages.setLanguageConfiguration('gaddis', {
            indentationRules: {
                increaseIndentPattern: /^\s*(While|Do|For|Select|Case|Default|If|Else|Module|Function|Class).*$/,
                decreaseIndentPattern: /^\s*((End|Until|Else|Case)\b)/
            }
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
                    stopOnEntry: true
                }).then(v => console.log(v));
            }
        }),
        vscode.commands.registerCommand('extension.gaddis.getProgramName', config => {
            return vscode.window.showInputBox({
                placeHolder: "Please enter the name of a gaddis file in the workspace folder",
                value: "program.gad"
            });
        }),

        vscode.debug.registerDebugConfigurationProvider('gaddis', new GaddisConfigurationProvider()),
        vscode.debug.registerDebugConfigurationProvider('gaddis', {
            provideDebugConfigurations(folder: WorkspaceFolder | undefined): ProviderResult<DebugConfiguration[]> {
                return [
                    {
                        name: "Dynamic Launch",
                        request: "launch",
                        type: "gaddis",
                        program: "${file}"
                    }
                ];
            }
        }, vscode.DebugConfigurationProviderTriggerKind.Dynamic),

        vscode.tasks.registerTaskProvider('gaddis.run', new GaddisRunProvider()),
        vscode.tasks.registerTaskProvider('gaddis.test', new GaddisTestProvider()),

        vscode.commands.registerCommand('gaddis.runTask', (fileUri: vscode.Uri) => {
            if (fileUri) {
                vscode.tasks.executeTask(makeTask('run', fileUri));
                return;
            }
            vscode.tasks.fetchTasks({ type: 'gaddis.run' }).then((tasks) => {
                if (tasks && tasks.length > 0) {
                    vscode.tasks.executeTask(tasks[0]);
                }
            });
        }),
        vscode.commands.registerCommand('gaddis.testTask', (fileUri: vscode.Uri) => {
            if (fileUri) {
                vscode.tasks.executeTask(makeTask('test', fileUri));
                return;
            }
            vscode.tasks.fetchTasks({ type: 'gaddis.test' }).then((tasks) => {
                if (tasks && tasks.length > 0) {
                    vscode.tasks.executeTask(tasks[0]);
                }
            });
        }),
    );

    activateDiagnostics(context);
    activateFormat(context);
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