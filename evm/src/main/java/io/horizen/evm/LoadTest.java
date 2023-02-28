package io.horizen.evm;

import com.google.common.base.Strings;

import java.math.BigInteger;
import java.util.Random;
import java.util.TimerTask;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

public class LoadTest {

    public static AtomicInteger counter = new AtomicInteger(0);

    public static class StatusTask extends TimerTask {
        private static final long BYTE_TO_MB_CONVERSION_VALUE = 1024 * 1024;

        private static long getCurrentlyUsedMemory() {
            System.gc();
            return (Runtime.getRuntime().totalMemory() - Runtime.getRuntime().freeMemory()) / BYTE_TO_MB_CONVERSION_VALUE;
        }

        @Override
        public void run() {
            System.out.printf("used memory: %d MB iterations: %d%n", getCurrentlyUsedMemory(), counter.get());
            counter.incrementAndGet();
        }
    }

    public static class LoadTask extends TimerTask {
        @Override
        public void run() {
            try {
                blockHashCallback();
                counter.incrementAndGet();
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }
    }

    public static void main(String[] args) {
        var N = 32;
        System.out.printf("starting bechmark tasks: %d%n", N);
        var executor = Executors.newScheduledThreadPool(N);
        try {
            for (int i = 0; i < N; i++) {
                executor.scheduleAtFixedRate(new LoadTask(), 1000, 10, TimeUnit.MILLISECONDS);
            }
            executor.scheduleAtFixedRate(new StatusTask(), 0, 1000, TimeUnit.MILLISECONDS);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
        System.out.println("done");
    }

    private static final Address addr1 = new Address("0x1234561234561234561234561234561234561230");
    private static final Address addr2 = new Address("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b");
    private static final BigInteger gasLimit = BigInteger.valueOf(200000);
    private static final BigInteger v10m = BigInteger.valueOf(10000000);
    private static final BigInteger v5m = BigInteger.valueOf(5000000);
    private static final BigInteger gasPrice = BigInteger.valueOf(10);

    private static void blockHashCallback() throws Exception {
        final var contractCode = Converter.fromHexString(
            "608060405234801561001057600080fd5b50610157806100206000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c8063557ed1ba11610066578063557ed1ba146100bf578063564b81ef146100c55780639663f88f146100cb578063aacc5a17146100d3578063d1a82a9d146100d957600080fd5b806315e812ad146100985780631a93d1c3146100ad57806342cbb15c146100b3578063455259cb146100b9575b600080fd5b485b6040519081526020015b60405180910390f35b4561009a565b4361009a565b3a61009a565b4261009a565b4661009a565b61009a6100e7565b4461009a565b6040514181526020016100a4565b60006100f46001436100fa565b40905090565b8181038181111561011b57634e487b7160e01b600052601160045260246000fd5b9291505056fea2646970667358221220a629106cbdbc0017022eedc70f72757902db9dc7881e188747a544aaa638345d64736f6c63430008120033");
        final var funcBlockHash = Converter.fromHexString("9663f88f");
        final var blockHashBytes = new byte[Hash.LENGTH];
        final var rand = new Random();
        rand.nextBytes(blockHashBytes);
        final var blockHash = new Hash(blockHashBytes);
        final var height = BigInteger.valueOf(rand.nextInt(999) + 1);

        class BlockHashGetter extends BlockHashCallback {
            @Override
            protected Hash getBlockHash(BigInteger blockNumber) {
//                System.out.println("block hash callback called with: blockNumber=" + blockNumber.toString());
                return blockHash;
            }
        }

        // TODO: try LevelDB instead of memory database
        try (
            var db = new MemoryDatabase();
            var statedb = new StateDB(db, Hash.ZERO);
            var blockHashGetter = new BlockHashGetter()
        ) {
            // deploy OpCode test contract
            var createResult = Evm.Apply(statedb, addr1, null, null, contractCode, gasLimit, gasPrice, null, null);
            var contractAddress = createResult.contractAddress;

            // setup context
            var context = new EvmContext();
            context.blockNumber = height;
            context.blockHashCallback = blockHashGetter;

            // call getBlockHash() function on the contract
            var result = Evm.Apply(
                statedb, addr1, contractAddress, null, funcBlockHash, gasLimit, gasPrice, context, null);
            if (!Strings.isNullOrEmpty(result.evmError) || result.reverted) {
                throw new RuntimeException(String.format(
                    "EVM execution error: %s (%s)", result.evmError, Converter.toHexString(result.returnData)
                ));
            }
            var returnedHash = new Hash(result.returnData);
//            System.out.printf("return value: %s%n", returnedHash);
            if (!blockHash.equals(returnedHash)) {
                throw new RuntimeException("bad block hash received");
            }
        }
    }
}
