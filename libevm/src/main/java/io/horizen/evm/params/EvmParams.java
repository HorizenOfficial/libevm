package io.horizen.evm.params;

import io.horizen.evm.Address;
import io.horizen.evm.EvmContext;

import java.math.BigInteger;

public class EvmParams extends HandleParams {
    public final Address from;
    public final Address to;
    public final BigInteger value;
    public final byte[] input;
    public final BigInteger availableGas; // uint64
    public final BigInteger gasPrice;
    public final EvmContext context;

    public EvmParams(
        int handle,
        Address from,
        Address to,
        BigInteger value,
        byte[] input,
        BigInteger availableGas,
        BigInteger gasPrice,
        EvmContext context
    ) {
        super(handle);
        this.from = from;
        this.to = to;
        this.value = value;
        this.input = input;
        this.availableGas = availableGas;
        this.gasPrice = gasPrice;
        this.context = context;
    }
}
