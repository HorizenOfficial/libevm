// SPDX-License-Identifier: MIT

pragma solidity >=0.7.0 <0.9.0;

import "./BaseNativeInterface.sol";

// This contract is used for testing interoperability between Solidity smart contracts and native smart contracts.
// Each function tests a possible interaction between the 2 contract types (e.g. call, static call, solidity interface...).
// It invokes functions of the dummy native contract defined in sdk/src/test/scala/io/horizen/account/state/ContractInteropCallTest.scala.
contract NativeCaller {

    BaseNativeInterface nativeContract = BaseNativeInterface(0x00000000000000000000000000000000DeaDBeef);

    uint256 private constant CALL_GAS_LIMIT = 25000;


    //This function tests that calling a readonly method on a Native Smart Contract using staticcall works
    // and that calling then a readwrite method works again, so the statedb is readwrite again.
    function testStaticCallOnReadonlyMethod() public returns (uint32) {
        address contractAddr = address(nativeContract);
        (bool success, bytes memory result) = contractAddr.staticcall{gas:CALL_GAS_LIMIT}(
            abi.encodeWithSignature("retrieve()")
        );
        (uint32 a) = abi.decode(result, (uint32));
        //Check that statedb is readwrite again
        (bool success2, bytes memory result2) = contractAddr.call{gas:CALL_GAS_LIMIT}(
            abi.encodeWithSignature("inc()")
        );
        require(success2, "call should work");
        return a;
    }

    //This function tests that calling a readwrite method on a Native Smart Contract using staticcall fails.
    //It tests also that calling the same method without staticcall works.         
    function testStaticCallOnReadwriteMethod() public returns (uint32) {
        address contractAddr = address(nativeContract);
        //This should fail
        (bool success, bytes memory result) = contractAddr.staticcall{gas:CALL_GAS_LIMIT}(
            abi.encodeWithSignature("inc()")
        );
        require(!success, "Staticcall should fail");
        //This should work instead.
        (bool success2, bytes memory result2) = contractAddr.call{gas:CALL_GAS_LIMIT}(
            abi.encodeWithSignature("inc()")
        );
        require(success2, "call should work");
        return abi.decode(result2, (uint32));
    }

    //This function calls a readwrite method on a Native Smart Contract using a contract call.
    // It should fail because the Solidity interface describing the Native Smart Contract defines the method as view,
    // even if it actually is readwrite. Using the contract interface, the tx should be automatically reverted.
    function testStaticCallOnReadwriteMethodContractCall() public returns (uint32) {
        return nativeContract.inc{gas: CALL_GAS_LIMIT}();
    }

    //This function is used to test staticcall with nested calls (native => evm => native).
    function testNestedCalls() public returns (uint32) {
        address contractAddr = address(nativeContract);
        (bool success, bytes memory result) = contractAddr.call{gas:CALL_GAS_LIMIT}(
            abi.encodeWithSignature("inc()")
        );
	require(success);
	return abi.decode(result, (uint32));
    }

}
