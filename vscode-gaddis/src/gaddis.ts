import * as vscode from 'vscode';
import { CancellationToken, DebugConfiguration, ProviderResult, WorkspaceFolder } from 'vscode';
import { GaddisRunProvider } from "./run_provider";
import { GaddisTestProvider } from "./test_provider";
import { makeTask } from './task';
import { activateDiagnostics } from './diagnostics';
import { activateFormat } from './format';
import { activateDebug } from './debug';

export function activate(context: vscode.ExtensionContext) {
    const subs = context.subscriptions;

    subs.push(
        vscode.languages.setLanguageConfiguration('gaddis', {
            indentationRules: {
                increaseIndentPattern: /^\s*(While|Do|For|Select|Case|Default|If|Else|Module|Function|Class).*$/,
                decreaseIndentPattern: /^\s*((End|Until|Else|Case)\b)/
            }
        }),

        vscode.tasks.registerTaskProvider('gaddis.run', new GaddisRunProvider()),
        vscode.commands.registerCommand('extension.gaddis.runTask', (fileUri: vscode.Uri) => {
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

        vscode.tasks.registerTaskProvider('gaddis.test', new GaddisTestProvider()),
        vscode.commands.registerCommand('extension.gaddis.testTask', (fileUri: vscode.Uri) => {
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
    activateDebug(context);
    activateFormat(context);
}
