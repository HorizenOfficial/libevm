package io.horizen.evm.params;

import io.horizen.evm.Address;

import java.math.BigInteger;

public class TracerEnterParams extends TracerParams {
    public final String opCode;
    public final Address from;
    public final Address to;
    public final byte[] input;
    public final BigInteger gas;
    public final BigInteger value;

    public TracerEnterParams(
        int tracerHandle,
        String opCode,
        Address from,
        Address to,
        byte[] input,
        BigInteger gas,
        BigInteger value
    ) {
        super(tracerHandle);
        this.opCode = opCode;
        this.from = from;
        this.to = to;
        this.input = input;
        this.gas = gas;
        this.value = value;
    }
}
