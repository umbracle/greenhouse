// SPDX-License-Identifier: MIT
pragma solidity >= 0.4.22 <0.9.0;

library testify {
	function assertEq(uint p0, uint p1) internal pure {
		if (p0 != p1) {
            revert();
        }
	}
}