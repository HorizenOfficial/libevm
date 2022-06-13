package com.horizen.evm.interop;

import com.horizen.evm.utils.Address;

import java.math.BigInteger;

public class EvmParams extends HandleParams {
    public Address from;
    public Address to;
    public BigInteger value;
    public byte[] input;
    public BigInteger nonce; // uint64
    public BigInteger gasLimit; // uint64
    public BigInteger gasPrice;

    public EvmContext context;

    public EvmParams() {
    }

    public EvmParams(int handle, byte[] from, byte[] to, BigInteger value, byte[] input, BigInteger nonce, BigInteger gasLimit, BigInteger gasPrice) {
        super(handle);
        this.from = Address.FromBytes(from);
        this.to = Address.FromBytes(to);
        this.value = value;
        this.input = input;
        this.nonce = nonce;
        this.gasLimit = gasLimit;
        this.gasPrice = gasPrice;
    }
}