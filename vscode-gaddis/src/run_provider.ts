import * as vscode from 'vscode';
import { makeTask } from "./task";

export class GaddisRunProvider implements vscode.TaskProvider {

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
        return [makeTask('run', editor.document.uri)]

    }
}
