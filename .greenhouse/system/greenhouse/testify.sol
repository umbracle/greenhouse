// SPDX-License-Identifier: MIT
pragma solidity >= 0.4.22 <0.9.0;

library testify {
    /*
	address constant CONSOLE_ADDRESS = address(0x000000000000000000636F6e736F6c652e6c6f67);

	function _sendLogPayload(bytes memory payload) private view {
		uint256 payloadLength = payload.length;
		address consoleAddress = CONSOLE_ADDRESS;
		assembly {
			let payloadStart := add(payload, 32)
			let r := staticcall(gas(), consoleAddress, payloadStart, payloadLength, 0, 0)
		}
	}
    */
    
	function assertEq(uint p0, uint p1) internal pure {
		if (p0 != p1) {
            revert();
        }
	}
}