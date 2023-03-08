package io.horizen.evm;

import java.util.Arrays;
import java.util.Random;

public class LibEvmTestBase {
    private static final Random rand = new Random();

    protected static byte[] bytes(String hex) {
        return Converter.fromHexString(hex);
    }

    protected static byte[] concat(byte[] a, byte[] b) {
        var merged = Arrays.copyOf(a, a.length + b.length);
        System.arraycopy(b, 0, merged, a.length, b.length);
        return merged;
    }

    protected static Hash padToHash(byte[] bytes) {
        var padded = new byte[Hash.LENGTH];
        System.arraycopy(bytes, 0, padded, padded.length - bytes.length, bytes.length);
        return new Hash(padded);
    }

    protected static Hash randomHash() {
        var bytes = new byte[Hash.LENGTH];
        rand.nextBytes(bytes);
        return new Hash(bytes);
    }
}
