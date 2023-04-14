package io.horizen.evm;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.math.BigInteger;

public abstract class BlockHashCallback extends LibEvmCallback {
    private static final Logger logger = LogManager.getLogger();

    protected abstract Hash getBlockHash(BigInteger blockNumber);

    @Override
    public String invoke(String args) {
        logger.debug("received block hash callback");
        try {
            var blockNumber = Converter.fromJson(args, BigInteger.class);
            return Converter.toJson(getBlockHash(blockNumber));
        } catch (Exception e) {
            // note: make sure we do not throw any exception here because this callback is called by native code
            // for diagnostics we log the exception here
            logger.warn("received invalid block hash collback", e);
        }
        return Converter.toJson(Hash.ZERO);
    }
}

