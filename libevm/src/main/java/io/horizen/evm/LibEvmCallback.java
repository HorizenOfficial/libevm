package io.horizen.evm;

import com.fasterxml.jackson.annotation.JsonValue;

/**
 * Base class to be used when passing callbacks to libevm. Can and should be used when passing parameter objects via the
 * LibEvm.invoke() JSON interface. When a parameter derived from this type is passed to libevm it will be serialized as
 * an integer handle which maps to the instance of the callback. Callback handles need to be disposed when not in use
 * anymore, call close() - or better - use the try-with-resources syntax.
 */
abstract class LibEvmCallback implements AutoCloseable {

    @JsonValue
    public final int handle;

    protected LibEvmCallback() {
        // acquire a callback handle on instantiation
        handle = CallbackRegistry.register(this);
    }

    @Override
    public void close() {
        // release the callback handle on close
        CallbackRegistry.unregister(handle, this);
    }

    public abstract String invoke(String args);
}
