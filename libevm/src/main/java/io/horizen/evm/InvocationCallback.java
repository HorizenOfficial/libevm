package io.horizen.evm;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

public abstract class InvocationCallback extends LibEvmCallback {
    private static final Logger logger = LogManager.getLogger();

    protected abstract InvocationResult execute(Invocation args);

    @Override
    public String invoke(String args) {
        logger.debug("received external contract callback");
        try {
            var invocation = Converter.fromJson(args, Invocation.class);
            return Converter.toJson(execute(invocation));
        } catch (Exception e) {
            // note: make sure we do not throw any exception here because this callback is called by native code
            // for diagnostics we log the exception here
            logger.warn("received invalid external contract callback", e);
        }
        return null;
    }
}
