/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */

import { exec } from 'child_process';
import * as path from 'path';
import { workspace, ExtensionContext } from 'vscode';

import * as vscode from 'vscode';
import * as net from 'net';

import { 
	WorkspaceFolder, 
	DebugConfiguration, 
	ProviderResult, 
	CancellationToken
} from 'vscode';

import {
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	TransportKind,
	Executable,
	StreamInfo,
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {

	const factory = new MockDebugAdapterServerDescriptorFactory()
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
	const executable: Executable = {
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
        const result: StreamInfo = {
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
	const clientOptions: LanguageClientOptions = {
		// Register the server for plain text documents
		documentSelector: [{ scheme: 'file', language: 'solidity' }],
		synchronize: {
			// Notify the server about file changes to '.clientrc files contained in the workspace
			fileEvents: workspace.createFileSystemWatcher('**/.clientrc')
		}
	};

	// Create the language client and start the client.
	client = new LanguageClient(
		'languageServerExample',
		'Language Server Example',
		serverOptions,
		clientOptions
	);

	// Start the client. This will also launch the server
	client.start();
}

export function deactivate(): Thenable<void> | undefined {
	if (!client) {
		return undefined;
	}
	return client.stop();
}


class MockConfigurationProvider implements vscode.DebugConfigurationProvider {

	/**
	 * Massage a debug configuration just before a debug session is being launched,
	 * e.g. add all missing attributes to the debug configuration.
	 */
	resolveDebugConfiguration(folder: WorkspaceFolder | undefined, config: DebugConfiguration, token?: CancellationToken): ProviderResult<DebugConfiguration> {

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
				return undefined;	// abort launch
			});
		}
		
		return config;
	}
}

class MockDebugAdapterServerDescriptorFactory implements vscode.DebugAdapterDescriptorFactory {

	//private server?: Net.Server;

	createDebugAdapterDescriptor(session: vscode.DebugSession, executable: vscode.DebugAdapterExecutable | undefined): vscode.ProviderResult<vscode.DebugAdapterDescriptor> {

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
