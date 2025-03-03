import * as vscode from 'vscode';
import * as path from 'path';

export class GaddisRunProvider implements vscode.TaskProvider {
    private gaddisExecutablePath: string;

    constructor(gaddisExecutablePath: string) {
        this.gaddisExecutablePath = gaddisExecutablePath;
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

        const filePath = editor.document.uri.fsPath;
        const fileName = path.basename(filePath);

        const task = new vscode.Task(
            {type: 'gaddis'},
            vscode.TaskScope.Workspace,
            `Run ${fileName}`,
            'Gaddis',
            new vscode.ShellExecution(`${this.gaddisExecutablePath} run "${filePath}"`)
        );

        task.group = vscode.TaskGroup.Build;
        task.presentationOptions.reveal = vscode.TaskRevealKind.Always;
        task.presentationOptions.panel = vscode.TaskPanelKind.New;

        return [task];
    }
}
