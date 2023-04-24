package io.horizen.evm;

import java.math.BigInteger;

public class EvmContext {
    public BigInteger chainID;
    public Address coinbase;
    public BigInteger gasLimit;
    public BigInteger gasPrice;
    public BigInteger blockNumber;
    public BigInteger time;
    public BigInteger baseFee;
    public Hash random;
    public BlockHashCallback blockHashCallback;
    public Address[] externalContracts;
    public InvocationCallback externalCallback;
    public Tracer tracer;
}
