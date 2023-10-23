package io.horizen.evm;

import io.horizen.evm.params.EvmParams;
import io.horizen.evm.results.InvocationResult;

public final class Evm {
    private Evm() { }

    public static InvocationResult Apply(ResourceHandle stateDBHandle, Invocation invocation, EvmContext context) {
        var params = new EvmParams(stateDBHandle.handle, invocation, context);
        return LibEvm.invoke("EvmApply", params, InvocationResult.class);
    }
}
