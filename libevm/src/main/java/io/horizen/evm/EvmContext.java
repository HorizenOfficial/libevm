package io.horizen.evm;

import java.math.BigInteger;

public class EvmContext {
    public final BigInteger chainID;
    public final Address coinbase;
    public final BigInteger gasLimit;
    public final BigInteger gasPrice;
    public final BigInteger blockNumber;
    public final BigInteger time;
    public final BigInteger baseFee;
    public final Hash random;
    public final ForkRules rules;
    private BlockHashCallback blockHashCallback;
    private Address[] externalContracts;
    private InvocationCallback externalCallback;
    private Tracer tracer;
    private int initialDepth;


    public EvmContext(BigInteger chainID,
                      Address coinbase,
                      BigInteger gasLimit,
                      BigInteger gasPrice,
                      BigInteger blockNumber,
                      BigInteger time,
                      BigInteger baseFee,
                      Hash random,
                      ForkRules rules) {
        this.chainID = chainID;
        this.coinbase = coinbase;
        this.gasLimit = gasLimit;
        this.gasPrice = gasPrice;
        this.blockNumber = blockNumber;
        this.time = time;
        this.baseFee = baseFee;
        this.random = random;
        this.rules = rules;
    }

    //This constructor is just for testing purposes
    EvmContext() {
        chainID = BigInteger.ZERO;
        coinbase = Address.ZERO;
        gasLimit = BigInteger.ZERO;
        gasPrice = BigInteger.ZERO;
        blockNumber = BigInteger.ZERO;
        time = BigInteger.ZERO;
        baseFee = BigInteger.ZERO;
        random = Hash.ZERO;
        rules = new ForkRules(false);
    }

    public BlockHashCallback getBlockHashCallback() {
        return blockHashCallback;
    }

    public void setBlockHashCallback(BlockHashCallback blockHashCallback) {
        this.blockHashCallback = blockHashCallback;
    }

    public Address[] getExternalContracts() {
        return externalContracts;
    }

    public void setExternalContracts(Address[] externalContracts) {
        this.externalContracts = externalContracts;
    }

    public InvocationCallback getExternalCallback() {
        return externalCallback;
    }

    public void setExternalCallback(InvocationCallback externalCallback) {
        this.externalCallback = externalCallback;
    }

    public Tracer getTracer() {
        return tracer;
    }

    public void setTracer(Tracer tracer) {
        this.tracer = tracer;
    }

    public int getInitialDepth() {
        return initialDepth;
    }

    public void setInitialDepth(int initialDepth) {
        this.initialDepth = initialDepth;
    }
}
