package io.horizen.evm.params;

import io.horizen.evm.Address;
import io.horizen.evm.ForkRules;

public class AccessParams extends AccountParams {
    public final Address destination;
    public final Address coinbase;
    public final ForkRules rules;

    public AccessParams(int handle, Address sender, Address destination, Address coinbase, ForkRules rules) {
        super(handle, sender);
        this.destination = destination;
        this.coinbase = coinbase;
        this.rules = rules;
    }
}
