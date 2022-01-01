//SPDX-License-Identifier: Unlicense
pragma solidity >=0.4.22 <0.6.0;

import "./B.sol";

contract B {
    string private greeting2;
}

contract Greeter is B, C {
    string private greeting;

    function greet() public view returns (string memory) {
        return greeting;
    }

    function setGreeting(string memory _greeting) public {
        greeting = _greeting;
    }
}
