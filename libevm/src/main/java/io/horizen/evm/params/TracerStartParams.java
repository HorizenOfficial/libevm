package io.horizen.evm.params;

import io.horizen.evm.Address;
import io.horizen.evm.EvmContext;

import java.math.BigInteger;

public class TracerStartParams extends TracerParams {
    public final int stateDB;
    public final EvmContext context;
    public final Address from;
    public final Address to;
    public final boolean create;
    public final byte[] input;
    public final BigInteger gas;
    public final BigInteger value;

    public TracerStartParams(
        int tracerHandle,
        int stateDB,
        EvmContext context,
        Address from,
        Address to,
        boolean create,
        byte[] input,
        BigInteger gas,
        BigInteger value
    ) {
        super(tracerHandle);
        this.stateDB = stateDB;
        this.context = context;
        this.from = from;
        this.to = to;
        this.create = create;
        this.input = input;
        this.gas = gas;
        this.value = value;
    }
}
