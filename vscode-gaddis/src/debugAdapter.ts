import * as child_process from 'child_process';
import { gaddisCmd } from "./platform";

const debugAdapterProcess = child_process.spawn(gaddisCmd, ["debug"], { stdio: 'inherit' });

debugAdapterProcess.on('error', (err) => {
    console.error('Failed to start debug adapter:', err);
    process.exit(1);
});

debugAdapterProcess.on('exit', (code) => {
    console.log(`Debug adapter exited with code ${code}`);
});

// Terminate the child process when the parent process exits
process.on('exit', () => {
    if (debugAdapterProcess && !debugAdapterProcess.killed) {
        debugAdapterProcess.kill('SIGTERM'); // Or 'SIGKILL' if necessary
    }
});

// Handle SIGINT (Ctrl+C) and SIGTERM gracefully
process.on('SIGINT', () => {
    process.exit(0);
});

process.on('SIGTERM', () => {
    process.exit(0);
});
