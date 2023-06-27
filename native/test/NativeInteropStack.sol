// SPDX-License-Identifier: MIT

pragma solidity >=0.7.0 <0.9.0;

contract NativeInteropStackContract {
    function loop(address target, uint32 counter) public returns (uint32) {
        try NativeInteropStackContract(target).loop(address(this), counter + 1) returns (uint32 returnedCounter) {
            return returnedCounter;
        } catch {
            return counter;
        }
    }
}
