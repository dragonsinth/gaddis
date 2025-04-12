import * as vscode from 'vscode';
import * as child_process from 'child_process';
import { gaddisCmd } from './platform';

export let dapServerPort: number = 0;

export function activateDapServer(context: vscode.ExtensionContext) {
    const outputChannel = vscode.window.createOutputChannel('Gaddis Debug Server', { log: true });
    let serverProcess: child_process.ChildProcessWithoutNullStreams | null = null;
    let shouldStop = false;

    function doRunServer() {
        serverProcess = child_process.spawn(gaddisCmd, ["-port", "0", "debug"], { stdio: 'pipe' });

        serverProcess.stdout.on('data', streamToLineAdapter(line => {
            outputChannel.trace(line);
        }));

        serverProcess.stderr.on('data', streamToLineAdapter(line => {
            if (dapServerPort == 0) {
                dapServerPort = parsePort(line);
            }
            outputChannel.info(line);
        }));

        serverProcess.on('exit', (code: any) => {
            outputChannel.appendLine(`Debug server exited with code ${code}; restarting.`);
            serverProcess = null;
            dapServerPort = 0;
            if (!shouldStop) {
                doRunServer();
            }
        });
    }

    doRunServer();

    function dispose() {
        shouldStop = true;
        if (serverProcess) {
            serverProcess.kill('SIGTERM');
        }
    }

    context.subscriptions.push(
        { dispose },
        outputChannel,
    );
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
