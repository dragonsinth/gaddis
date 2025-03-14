import * as vscode from 'vscode';
import { activateDebug } from './debug';
import { activateDiagnostics } from './diagnostics';
import { activateFormat } from './format';
import { activateDapServer } from "./dapServer";

export function activate(context: vscode.ExtensionContext) {
    context.subscriptions.push(
        vscode.languages.setLanguageConfiguration('gaddis', {
            indentationRules: {
                increaseIndentPattern: /^\s*(While|Do|For|Select|Case|Default|If|Else|Module|Function|Class).*$/,
                decreaseIndentPattern: /^\s*((End|Until|Else|Case)\b)/
            }
        }),
    );
    activateDapServer(context);
    activateDebug(context);
    activateDiagnostics(context);
    activateFormat(context);
}
