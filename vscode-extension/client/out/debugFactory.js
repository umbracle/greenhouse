"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.DebugAdapterDescriptorFactory = void 0;
const vscode = require("vscode");
class DebugAdapterDescriptorFactory {
    createDebugAdapterDescriptor(_session, executable) {
        console.log("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX");
        return new vscode.DebugAdapterServer(4569);
    }
    dispose() {
    }
}
exports.DebugAdapterDescriptorFactory = DebugAdapterDescriptorFactory;
//# sourceMappingURL=debugFactory.js.map