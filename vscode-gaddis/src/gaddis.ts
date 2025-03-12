import * as vscode from 'vscode';
import * as path from 'path';
import { activateDebug } from './debug';
import { activateDiagnostics } from './diagnostics';
import { activateFormat } from './format';
import { gaddisCmd } from './platform';

export function activate(context: vscode.ExtensionContext) {
    const subs = context.subscriptions;

    subs.push(
        vscode.languages.setLanguageConfiguration('gaddis', {
            indentationRules: {
                increaseIndentPattern: /^\s*(While|Do|For|Select|Case|Default|If|Else|Module|Function|Class).*$/,
                decreaseIndentPattern: /^\s*((End|Until|Else|Case)\b)/
            }
        }),

        vscode.tasks.registerTaskProvider('extension.gaddis.test', new GaddisTestProvider()),
        vscode.commands.registerCommand('extension.gaddis.testTask', (fileUri: vscode.Uri) => {
            if (fileUri) {
                vscode.tasks.executeTask(makeTask('test', fileUri));
                return;
            }
            vscode.tasks.fetchTasks({ type: 'extension.gaddis.test' }).then((tasks) => {
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


class GaddisTestProvider implements vscode.TaskProvider {

    constructor() {
    }

    resolveTask(task: vscode.Task, token: vscode.CancellationToken): vscode.ProviderResult<vscode.Task> {
        // Not needed for simple run tasks
        return task;
    }

    provideTasks(token: vscode.CancellationToken): vscode.ProviderResult<vscode.Task[]> {
        const editor = vscode.window.activeTextEditor;
        if (!editor || editor.document.languageId !== 'gaddis') {
            return [];
        }
        return [makeTask('test', editor.document.uri)]
    }
}

function makeTask(cmd: string, fileUri: vscode.Uri): vscode.Task {
    const filePath = fileUri.fsPath;
    const fileName = path.basename(filePath);

    const task = new vscode.Task(
        { type: `gaddis.${cmd}` },
        vscode.TaskScope.Workspace,
        `${cmd} ${fileName}`,
        'gaddis',
        new vscode.ShellExecution(`${gaddisCmd} ${cmd} "${filePath}"`)
    );

    task.group = vscode.TaskGroup.Test;
    task.presentationOptions.reveal = vscode.TaskRevealKind.Always;
    task.presentationOptions.panel = vscode.TaskPanelKind.New;
    return task
}
