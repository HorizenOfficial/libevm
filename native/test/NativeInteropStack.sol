// SPDX-License-Identifier: MIT

pragma solidity >=0.7.0 <0.9.0;

contract NativeInteropStack {
    function loop(address target, uint32 counter) public returns (uint32) {
        try NativeInteropStack(target).loop(address(this), counter + 1) returns (uint32 returnedCounter) {
            return returnedCounter;
        } catch {
            return counter;
        }
    }
}
