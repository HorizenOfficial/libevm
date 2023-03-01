package io.horizen.evm;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.util.HashMap;
import java.util.Map;

class CallbackRegistry {
    private static final Logger logger = LogManager.getLogger();

    private static final Map<Integer, LibEvmCallback> callbacks = new HashMap<>();

    private CallbackRegistry() {
        // prevent instantiation, this class only has static members
    }

    static synchronized int register(LibEvmCallback callback) {
        // note: with N items in the map this will iterate for N+1 times, hence it should always find an unused handle
        for (int handle = 0; handle <= callbacks.size(); handle++) {
            if (!callbacks.containsKey(handle)) {
                callbacks.put(handle, callback);
                logger.trace("registered callback with handle {}: {}", handle, callback);
                return handle;
            }
        }
        throw new IllegalStateException("too many callback handles");
    }

    static synchronized void unregister(int handle, LibEvmCallback callback) {
        if (!callbacks.remove(handle, callback)) {
            logger.warn("already unregistered callback with handle {}: {}", handle, callback);
            return;
        }
        logger.trace("unregistered callback with handle {}: {}", handle, callback);
    }

    static synchronized String invoke(int handle, String args) {
        if (!callbacks.containsKey(handle)) {
            logger.warn("received callback with invalid handle: {}", handle);
            return null;
        }
        return callbacks.get(handle).invoke(args);
    }
}
