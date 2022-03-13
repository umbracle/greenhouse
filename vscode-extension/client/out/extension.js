"use strict";
/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */
Object.defineProperty(exports, "__esModule", { value: true });
exports.deactivate = exports.activate = void 0;
const path = require("path");
const vscode_1 = require("vscode");
const vscode = require("vscode");
const net = require("net");
const node_1 = require("vscode-languageclient/node");
let client;
function activate(context) {
    // The server is implemented in node
    const serverModule = context.asAbsolutePath(path.join('server', 'out', 'server.js'));
    // The debug options for the server
    // --inspect=6009: runs the server in Node's Inspector mode so VS Code can attach to the server for debugging
    const debugOptions = { execArgv: ['--nolazy', '--inspect=6009'] };
    console.log("YYYYYYYYYYYYYYYYYYYYYYYYYY");
    const factory = new MockDebugAdapterServerDescriptorFactory();
    context.subscriptions.push(vscode.debug.registerDebugAdapterDescriptorFactory('mock', factory));
    // register a configuration provider for 'mock' debug type
    const provider = new MockConfigurationProvider();
    context.subscriptions.push(vscode.debug.registerDebugConfigurationProvider('mock', provider));
    /*
    const executable: Executable = {
        command: "node",
        args: ["/home/ferran/workspace/vscode-extension-samples/lsp-sample/server/out/server.js", "--stdio"],
        options: {},
    };
    */
    const executable = {
        command: "/home/ferran/go/src/github.com/umbracle/greenhouse/main",
        args: ["lsp"],
        options: {},
    };
    const serverOptions = () => {
        const socket = new net.Socket();
        socket.connect({
            port: 4564,
            host: 'localhost',
        });
        const result = {
            writer: socket,
            reader: socket,
        };
        return Promise.resolve(result);
    };
    // If the extension is launched in debug mode then the debug server options are used
    // Otherwise the run options are used
    //const serverOptions: ServerOptions = {
    // run: { module: serverModule, transport: TransportKind.stdio },
    //run: executable,
    //debug: executable,
    /*
    debug: {
        module: serverModule,
        transport: TransportKind.ipc,
        options: debugOptions
    }
    */
    //};
    // Options to control the language client
    const clientOptions = {
        // Register the server for plain text documents
        documentSelector: [{ scheme: 'file', language: 'plaintext' }],
        synchronize: {
            // Notify the server about file changes to '.clientrc files contained in the workspace
            fileEvents: vscode_1.workspace.createFileSystemWatcher('**/.clientrc')
        }
    };
    // Create the language client and start the client.
    client = new node_1.LanguageClient('languageServerExample', 'Language Server Example', serverOptions, clientOptions);
    // Start the client. This will also launch the server
    client.start();
}
exports.activate = activate;
function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
exports.deactivate = deactivate;
class MockConfigurationProvider {
    /**
     * Massage a debug configuration just before a debug session is being launched,
     * e.g. add all missing attributes to the debug configuration.
     */
    resolveDebugConfiguration(folder, config, token) {
        // if launch.json is missing or empty
        if (!config.type && !config.request && !config.name) {
            config.type = 'mock';
            config.name = 'Launch';
            config.request = 'launch';
            config.program = '${file}';
            config.stopOnEntry = true;
        }
        if (!config.program) {
            return vscode.window.showInformationMessage("Cannot find a program to debug").then(_ => {
                return undefined; // abort launch
            });
        }
        return config;
    }
}
class MockDebugAdapterServerDescriptorFactory {
    //private server?: Net.Server;
    createDebugAdapterDescriptor(session, executable) {
        /*
        if (!this.server) {
            // start listening on a random port
            this.server = Net.createServer(socket => {
                const session = new MockDebugSession(workspaceFileAccessor);
                session.setRunAsServer(true);
                session.start(socket as NodeJS.ReadableStream, socket);
            }).listen(0);
        }
        */
        // make VS Code connect to debug server
        return new vscode.DebugAdapterServer(4569);
    }
    dispose() {
        /*
        if (this.server) {
            this.server.close();
        }
        */
    }
}
//# sourceMappingURL=extension.js.map