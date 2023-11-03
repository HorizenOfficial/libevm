package io.horizen.evm.params;

import io.horizen.evm.EvmContext;
import io.horizen.evm.Invocation;

public class EvmParams extends HandleParams {
    public final Invocation invocation;
    public final EvmContext context;

    public EvmParams(int handle, Invocation invocation, EvmContext context) {
        super(handle);
        this.invocation = invocation;
        this.context = context;
    }
}
