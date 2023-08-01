// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;


// contract address: 0x00000000000000000000000000000000deadbeef
interface BaseNativeInterface {

    function retrieve() external view returns (uint32);

    function inc() external view returns (uint32);
}
