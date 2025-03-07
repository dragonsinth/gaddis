import * as vscode from 'vscode';
import * as path from 'path';

type cmdType = 'run' | 'test'

export function makeTask(gaddisCmd: string, cmd: cmdType, fileUri: vscode.Uri): vscode.Task {
    const filePath = fileUri.fsPath;
    const fileName = path.basename(filePath);

    const task = new vscode.Task(
        { type: `gaddis.${cmd}` },
        vscode.TaskScope.Workspace,
        `${cmd} ${fileName}`,
        'Gaddis',
        new vscode.ShellExecution(`${gaddisCmd} ${cmd} "${filePath}"`)
    );

    task.group = vscode.TaskGroup.Test;
    task.presentationOptions.reveal = vscode.TaskRevealKind.Always;
    task.presentationOptions.panel = vscode.TaskPanelKind.New;
    return task
}
