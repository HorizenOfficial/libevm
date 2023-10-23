package io.horizen.evm;

import io.horizen.evm.results.InvocationResult;
import org.junit.Test;

import java.math.BigInteger;

import static org.junit.Assert.*;

public class EvmTest extends LibEvmTestBase {

    private final Address addr1 = new Address("0x1234561234561234561234561234561234561230");
    private final Address addr2 = new Address("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b");
    private final BigInteger gasLimit = BigInteger.valueOf(500000);
    private final BigInteger v10m = BigInteger.valueOf(10000000);
    private final BigInteger v5m = BigInteger.valueOf(5000000);

    private Invocation call(Address from, Address to, BigInteger value, byte[] input) {
        return new Invocation(from, to, value, input, gasLimit, false);
    }

    private Invocation create(Address from, byte[] input) {
        return new Invocation(from, null, null, input, gasLimit, false);
    }

    @Test
    public void evmApply() throws Exception {
        final var txHash = new Hash("0x4545454545454545454545454545454545454545454545454545454545454545");
        final var codeHash = new Hash("0xaa87aee0394326416058ef46b907882903f3646ef2a6d0d20f9e705b87c58c77");

        final var contractCode = bytes(
            "608060405234801561001057600080fd5b5060405161023638038061023683398101604081905261002f916100f6565b6000819055604051339060008051602061021683398151915290610073906020808252600c908201526b48656c6c6f20576f726c642160a01b604082015260600190565b60405180910390a2336001600160a01b03166000805160206102168339815191526040516100bf906020808252600a908201526948656c6c6f2045564d2160b01b604082015260600190565b60405180910390a26040517ffe1a3ad11e425db4b8e6af35d11c50118826a496df73006fc724cb27f2b9994690600090a15061010f565b60006020828403121561010857600080fd5b5051919050565b60f98061011d6000396000f3fe60806040526004361060305760003560e01c80632e64cec1146035578063371303c01460565780636057361d14606a575b600080fd5b348015604057600080fd5b5060005460405190815260200160405180910390f35b348015606157600080fd5b506068607a565b005b606860753660046086565b600055565b6000546075906001609e565b600060208284031215609757600080fd5b5035919050565b6000821982111560be57634e487b7160e01b600052601160045260246000fd5b50019056fea2646970667358221220769e4dd8320afae06d27e8e201c885728883af2ea321d02071c47704c1b3c24f64736f6c634300080e00330738f4da267a110d810e6e89fc59e46be6de0c37b1d5cd559b267dc3688e74e0");
        final var testValue = new Hash("0x00000000000000000000000000000000000000000000000000000000000015b3");

        final var funcStore = bytes("6057361d");
        final var funcRetrieve = bytes("2e64cec1");

        InvocationResult result;
        Address contractAddress;
        Hash modifiedStateRoot;
        byte[] calldata;

        try (var db = new MemoryDatabase()) {
            try (var statedb = new StateDB(db, Hash.ZERO)) {
                // test a simple value transfer
                statedb.addBalance(addr1, v10m);
                result = Evm.Apply(statedb, call(addr1, addr2, v5m, null), null);
                assertEquals("", result.executionError);
                assertEquals(v5m, statedb.getBalance(addr2));
                // gas fees should not have been deducted
                assertEquals(v5m, statedb.getBalance(addr1));
                // gas fees should not be moved to the coinbase address (which currently defaults to the zero-address)
                assertEquals(BigInteger.ZERO, statedb.getBalance(Address.ZERO));

                // test contract deployment
                calldata = concat(contractCode, Hash.ZERO.toBytes());
                statedb.setTxContext(txHash, 0);
                var createResult = Evm.Apply(statedb, create(addr2, calldata), null);
                assertEquals("", createResult.executionError);
                contractAddress = createResult.contractAddress;
                assertEquals(codeHash, statedb.getCodeHash(contractAddress));
                var logs = statedb.getLogs(txHash);
                assertEquals("should generate 3 log entries", 3, logs.length);
                for (var log : logs) {
                    assertEquals(log.address, createResult.contractAddress);
                }

                // call "store" function on the contract to set a value
                calldata = concat(funcStore, testValue.toBytes());
                result = Evm.Apply(statedb, call(addr2, contractAddress, null, calldata), null);
                assertEquals("", result.executionError);

                // use a tracer for the next call to verify it is used
                try (var tracer = new Tracer(new TraceOptions())) {
                    var context = new EvmContext();
                    context.setTracer(tracer);
                    // call "retrieve" on the contract to fetch the value we just set
                    result = Evm.Apply(statedb, call(addr2, contractAddress, null, funcRetrieve), context);
                    assertEquals("", result.executionError);
                    assertEquals(testValue, new Hash(result.returnData));
                    var trace = tracer.getResult().result;
                    assertNotNull(trace);
                    // verify that there is something in the structLogs array
                    assertTrue("unexpected tracer result", trace.withArray("structLogs").hasNonNull(0));
                }

                modifiedStateRoot = statedb.commit();
            }

            // reopen the state and retrieve a value
            try (var statedb = new StateDB(db, modifiedStateRoot)) {
                result = Evm.Apply(statedb, call(addr2, contractAddress, null, funcRetrieve), null);
                assertEquals("", result.executionError);
                assertEquals(testValue, new Hash(result.returnData));
            }
        }
    }

    @Test
    public void blockHashCallback() throws Exception {
        // compiled OpCodes.sol
        final var contractCode = bytes(
            "608060405234801561001057600080fd5b50610157806100206000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c8063557ed1ba11610066578063557ed1ba146100bf578063564b81ef146100c55780639663f88f146100cb578063aacc5a17146100d3578063d1a82a9d146100d957600080fd5b806315e812ad146100985780631a93d1c3146100ad57806342cbb15c146100b3578063455259cb146100b9575b600080fd5b485b6040519081526020015b60405180910390f35b4561009a565b4361009a565b3a61009a565b4261009a565b4661009a565b61009a6100e7565b4461009a565b6040514181526020016100a4565b60006100f46001436100fa565b40905090565b8181038181111561011b57634e487b7160e01b600052601160045260246000fd5b9291505056fea2646970667358221220a629106cbdbc0017022eedc70f72757902db9dc7881e188747a544aaa638345d64736f6c63430008120033");
        // signature for getBlockHash()
        final var funcBlockHash = bytes("9663f88f");
        final var blockHash = randomHash();
        final var height = BigInteger.valueOf(1234);

        class BlockHashGetter extends BlockHashCallback {
            private boolean throwIfCalled;

            public void enable() { throwIfCalled = true; }

            public void disable() { throwIfCalled = false; }

            @Override
            protected Hash getBlockHash(BigInteger blockNumber) {
                assertFalse("should not have been called", throwIfCalled);
                // getBlockHash() on the OpCode test contract should request the block hash for height - 1
                assertEquals("unexpected block hash requested", height.subtract(BigInteger.ONE), blockNumber);
                return blockHash;
            }
        }

        try (
            var db = new MemoryDatabase();
            var statedb = new StateDB(db, Hash.ZERO);
            var blockHashGetterA = new BlockHashGetter();
            var blockHashGetterB = new BlockHashGetter()
        ) {
            // deploy OpCode test contract
            var createResult = Evm.Apply(statedb, create(addr1, contractCode), null);
            assertEquals("", createResult.executionError);
            var contractAddress = createResult.contractAddress;

            // setup context
            var context = new EvmContext(BigInteger.ZERO,
                    Address.ZERO,
                    BigInteger.ZERO,
                    BigInteger.ZERO,
                    height,
                    BigInteger.ZERO,
                    BigInteger.ZERO,
                    Hash.ZERO) {
            };

            //context.blockNumber = height;
            context.setBlockHashCallback(blockHashGetterA);

            // throw if B is called
            blockHashGetterA.disable();
            blockHashGetterB.enable();

            // call getBlockHash() function on the contract
            var resultA = Evm.Apply(statedb, call(addr1, contractAddress, null, funcBlockHash), context);
            assertEquals("unexpected error message", "", resultA.executionError);
            assertEquals("unexpected block hash", blockHash, new Hash(resultA.returnData));

            // throw if A is called
            context.setBlockHashCallback(blockHashGetterB);
            blockHashGetterA.enable();
            blockHashGetterB.disable();

            // call getBlockHash() function on the contract
            var resultB = Evm.Apply(statedb, call(addr1, contractAddress, null, funcBlockHash), context);
            assertEquals("unexpected error message", "", resultB.executionError);
            assertEquals("unexpected block hash", blockHash, new Hash(resultB.returnData));
        }

        // sanity check for unregistering callbacks
        try (var blockHashGetterC = new BlockHashGetter()) {
            // handle 0 will always be used by the log callback
            // we released all other callbacks and created a new one here, so we expect the handle to be 1
            assertEquals("callback handles were not released", 1, blockHashGetterC.handle);
        }
    }

    @Test
    public void invocationCallback() throws Exception {
        // compiled NativeInterop.sol
        final var contractCode = bytes(
            "6080604052600080546001600160a01b031916692222222222222222222217905534801561002c57600080fd5b506107108061003c6000396000f3fe60806040526004361061004a5760003560e01c806324a084df1461004f57806367a7dbb414610064578063b63fc52914610079578063cb14b856146100a4578063e08b6262146100b9575b600080fd5b61006261005d3660046103d1565b6100ee565b005b34801561007057600080fd5b5061006261019d565b34801561008557600080fd5b5061008e610242565b60405161009b91906103fd565b60405180910390f35b3480156100b057600080fd5b506100626102c6565b3480156100c557600080fd5b506100d96100d436600461049a565b61031b565b60405163ffffffff909116815260200161009b565b600080836001600160a01b03163460405160006040518083038185875af1925050503d806000811461013c576040519150601f19603f3d011682016040523d82523d6000602084013e610141565b606091505b5091509150816101975760405162461bcd60e51b815260206004820152601860248201527f6661696c656420746f207472616e736665722076616c75650000000000000000604482015260640160405180910390fd5b50505050565b6000805460408051600481526024810182526020810180516001600160e01b031663f6ad3c2360e01b179052905183926001600160a01b031691612710916101e591906104d3565b6000604051808303818686f4925050503d8060008114610221576040519150601f19603f3d011682016040523d82523d6000602084013e610226565b606091505b50909250905081151560000361023e57805160208201fd5b5050565b606060008054906101000a90046001600160a01b03166001600160a01b031663f6ad3c236127106040518263ffffffff1660e01b81526004016000604051808303818786fa158015610298573d6000803e3d6000fd5b50505050506040513d6000823e601f3d908101601f191682016040526102c19190810190610572565b905090565b60008054604080517ff6ad3c23f0605b9ed84e6ad346e341d181873063303443c922270a3f389ee85e80825260048083019093526001600160a01b03909316939091602091839190829087612710f250505050565b60006001600160a01b03831663e08b626230610338856001610684565b6040516001600160e01b031960e085901b1681526001600160a01b03909216600483015263ffffffff1660248201526044016020604051808303816000875af19250505080156103a5575060408051601f3d908101601f191682019092526103a2918101906106b6565b60015b6103b05750806103b3565b90505b92915050565b6001600160a01b03811681146103ce57600080fd5b50565b600080604083850312156103e457600080fd5b82356103ef816103b9565b946020939093013593505050565b602080825282518282018190526000919060409081850190868401855b8281101561047b578151805185528681015187860152858101516001600160a01b031686860152606080820151908601526080808201519086015260a0908101516001600160f81b0319169085015260c0909301929085019060010161041a565b5091979650505050505050565b63ffffffff811681146103ce57600080fd5b600080604083850312156104ad57600080fd5b82356104b8816103b9565b915060208301356104c881610488565b809150509250929050565b6000825160005b818110156104f457602081860181015185830152016104da565b506000920191825250919050565b634e487b7160e01b600052604160045260246000fd5b60405160c0810167ffffffffffffffff8111828210171561053b5761053b610502565b60405290565b604051601f8201601f1916810167ffffffffffffffff8111828210171561056a5761056a610502565b604052919050565b6000602080838503121561058557600080fd5b825167ffffffffffffffff8082111561059d57600080fd5b818501915085601f8301126105b157600080fd5b8151818111156105c3576105c3610502565b6105d1848260051b01610541565b818152848101925060c09182028401850191888311156105f057600080fd5b938501935b828510156106785780858a03121561060d5760008081fd5b610615610518565b855181528686015187820152604080870151610630816103b9565b90820152606086810151908201526080808701519082015260a0808701516001600160f81b0319811681146106655760008081fd5b90820152845293840193928501926105f5565b50979650505050505050565b63ffffffff8181168382160190808211156106af57634e487b7160e01b600052601160045260246000fd5b5092915050565b6000602082840312156106c857600080fd5b81516106d381610488565b939250505056fea26469706673582212206fa1a6a9416df343205cf04e82be4d29a3914e50dccde29bd26f835048107a5264736f6c63430008140033");
        // signature for getForgerStakes() in NativeInterop.sol
        final var funcGetForgerStakes = bytes("b63fc529");
        // signature for getForgerStakesDelegateCall() in NativeInterop.sol
        final var funcGetForgerStakesDelegateCall = bytes("67a7dbb4");
        // signature for getAllForgersStakes() in ForgerStakes.sol
        final var funcGetAllForgerStakes = bytes("f6ad3c23");
        final var forgerStakesContractAddress = new Address("0x0000000000000000000022222222222222222222");
        final var mockedForgerStakesData = bytes(
            "0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000bc22971e6a19a3ddf28dafcbe3eaf261cfac0f3dc07f9cef79dfc94175d1eb8cc000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000039ef4eea2b229c11dba18392b7f12c83c89053b9ffc4065704d28337afdb9e216e264be8e1121f29b33a84860b78008cf8638ebe0bfce2073aa51bc5854f8f297e31b89358a29e27f95380b91268e68dc58d1d0a8000000000000000000000000000000000000000000000000000000000000000fe9436b7f4645cc5562e7f37996a11a63da703043df985ec23f9c5a642a288ff000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000002a789f245142753075c4c5a3c603b52d8f01d361b85bb1e0b3de12c8bb3b7f7cb04ee5f7cdec4de4bf0879536824cc84adf6d7d29e9f78162e4509a0df40e3d8525801c363a8d327ce0c9eb4da8a3615489aee1680000000000000000000000000000000000000000000000000000000000000000e19a05f03af33541c3219a734b0868f39b36493b3e8ba63c86c5659725d582c000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000551c0294ef40e7d4da0d0e62f77141cdb455dcc47b48711bc08c786c486520bb9eb69ba43ed04fc2cc617c24e85d757ffcaf9a971644ba05f4fcea365fabba5b1ee38f0a1db512f7178c3733e8680589cb4d7f350000000000000000000000000000000000000000000000000000000000000000aa27870083d34abec0abf7a0a5e39d2cb353ffb96c486af53868783147211f18000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000072b6b2ed1c4d951a5247206b771a2615ec25a255087a518986785e8746d114d6cbdea0fc9f048362b1b318b4d36f7d1d51596ac789c3617a58b703631dcff9ce96697e0b4e1c7ce9207db22edf110e5b29347c1580000000000000000000000000000000000000000000000000000000000000001ddda8086cbc1d0752e95651ba810b7a56441bf42a770010475f8d02774f7411000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000e36c4e40d357e4a6e4df7a05f5944264df867265c0905c48964a940880e8de76c84dddc8535d0878a693097b3e86f6509e649606b5c5916187e905d83ba75ebdc39d2d526d6402a4801677f325dd55e1df5d810d000000000000000000000000000000000000000000000000000000000000000086c07dadb86c083eb56e6e9ac3891afab10806b8221b50a5763ad56748b51113000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000002e0126cbcae91e490522f65925773d182339799674cac9966adc776d28720f7bfa744e7c725b7388734bb6801b98f3470aefdf3f0ce9e0bd37fb0abcc3dc392baeeb8ac3acd47ac73ae79429cd9f4807faec06330000000000000000000000000000000000000000000000000000000000000000fcbd1f41cf3f859f0f3c3de58d4227032400e9edc46feb81ad0d9790283be923000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000008bf4857b740943b005fbbee80719421ab5b3dc510ef5b3566378c969af6545b06706e380cba6ba6dc55d7c1a5ae0c914475100c7c7b855d86e93a45d0108026614c4a8e2acb201f7c4e0f26297e3fbde0cf57e268000000000000000000000000000000000000000000000000000000000000000703bf3e9519b03c8e97f29d5e4cce7e136f894a6fd5eb35020f193ffb77b838c000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000008a698b7535d6eeb6341e5c65e220c49f177f73e6cd00ba09b5d2b771c1f90399e183bf80e54886a390904be2e018d0d80b46b018d71700c48baf671f9ebd48895f555f32b7b0c20ee2a3dcb35290c9a412a2e03800000000000000000000000000000000000000000000000000000000000000006f64a8254e75c9dd9dba6499112c0fe99b7ad5f1d978ace402c3459d1d09891000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000030d617085728a81563dfd655c35020ab8b78ecec76f4fb6b460dbc0fdf3e17fb5b9888ab9ce2f2f009e2fb32ef95eff196b4326eacde2c72c65e185d9e04e6012ce0bc6165a17a4f95215d701982c8b8cb5d970800000000000000000000000000000000000000000000000000000000000000008d546419e013bafbaead980365902a624f724cd1611086ab55946ec72cbfe889000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000e089abfcb53895adcbc8825aba391ebd1ddd83ed12e73f518bd4314b5251d881394711433f348ebea66b817aade3c93ffb3650dc51c95d2cd6428da7189c3e7451e51906bde78fee932f3a61644c22ca8856f0320000000000000000000000000000000000000000000000000000000000000000fda9653e35779b6b8f5186548d9bd3e7f66f6bc662d530569104cc1ad6afe437000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000075413637bc6e1818b36c9acf56fac28765bc3ec8fa9471609398bc7520e8151a96409151433d426a094fc70e6f32b89ed854d12acd7a2250e51d6ef985275847e7c715e67b4967a44910d19950c19bf43578a3210000000000000000000000000000000000000000000000000000000000000000");

        try (
            var db = new MemoryDatabase();
            var statedb = new StateDB(db, Hash.ZERO)
        ) {
            // deploy NativeInterop test contract
            var createResult = Evm.Apply(statedb, create(addr1, contractCode), null);
            assertEquals("", createResult.executionError);
            var contractAddress = createResult.contractAddress;

            class NativeContractCallback extends InvocationCallback {
                @Override
                protected InvocationResult execute(ExternalInvocation args) {
                    assertEquals("expected call from deployed contract", contractAddress, args.caller);
                    assertEquals("expected call to forger stakes contract", forgerStakesContractAddress, args.callee);
                    assertEquals("expected call with no value", BigInteger.ZERO, args.value);
                    assertArrayEquals("expected call to GetForgerStakes()", funcGetAllForgerStakes, args.input);
                    assertEquals("expected call with 10k gas", BigInteger.valueOf(10000), args.gas);
                    assertTrue("expected read only flag to be set (STATICCALL)", args.readOnly);
                    assertEquals("unexpected call depth", 1, args.depth);
                    return new InvocationResult(mockedForgerStakesData, BigInteger.ZERO, "", false, null);
                }
            }
            try (var nativeContractCallback = new NativeContractCallback()) {
                // setup context
                var context = new EvmContext();
                context.setExternalContracts(new Address[] {forgerStakesContractAddress});
                context.setExternalCallback(nativeContractCallback);
                context.setInitialDepth(20);

                // call GetForgerStakes() function on the contract
                var callResult = Evm.Apply(
                    statedb,
                    new Invocation(addr1, contractAddress, null, funcGetForgerStakes, gasLimit, false),
                    context
                );
                assertEquals("unexpected error message", "", callResult.executionError);
                assertArrayEquals("unexpected forger stakes data", mockedForgerStakesData, callResult.returnData);

                // call GetForgerStakesDelegateCall() function on the contract,
                // this should fail because DELEGATECALL and CALLCODE to a native contract are not supported
                var delegateCallResult =
                    Evm.Apply(statedb, call(addr1, contractAddress, null, funcGetForgerStakesDelegateCall), context);
                assertEquals("expected all gas to be burned", BigInteger.ZERO, delegateCallResult.leftOverGas);
                assertFalse("unexpected revert, this should fail without revert", delegateCallResult.reverted);
                assertTrue(
                    "unexpected error message",
                    delegateCallResult.executionError.contains("unsupported call method")
                );
            }
        }
    }

    @Test
    public void insufficientBalanceTransfer() throws Exception {
        try (var db = new MemoryDatabase(); var statedb = new StateDB(db, Hash.ZERO)) {
            var result = Evm.Apply(statedb, call(addr1, addr2, v5m, null), null);
            assertEquals("unexpected error message", "insufficient balance for transfer", result.executionError);
            assertEquals("unexpected gas usage", gasLimit, result.leftOverGas);
        }
    }

    @Test
    public void brokenCodeExecution() throws Exception {
        final var input = bytes(
            "5234801561001057600080fd521683398151915290610073906020808252600c90820190565b60405180910390a2336001600160a01b03");
        try (var db = new MemoryDatabase(); var statedb = new StateDB(db, Hash.ZERO)) {
            statedb.setBalance(addr1, v5m);
            var result = Evm.Apply(statedb, create(addr1, input), null);
            assertTrue("unexpected error message", result.executionError.startsWith("stack underflow"));
            assertEquals("unexpected gas usage", BigInteger.ZERO, result.leftOverGas);
        }
    }

    @Test
    public void insufficientGasLimit() throws Exception {
        final var input = bytes(
            "608060405234801561001057600080fd5b50610157806100206000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c8063557ed1ba11610066578063557ed1ba146100bf578063564b81ef146100c55780639663f88f146100cb578063aacc5a17146100d3578063d1a82a9d146100d957600080fd5b806315e812ad146100985780631a93d1c3146100ad57806342cbb15c146100b3578063455259cb146100b9575b600080fd5b485b6040519081526020015b60405180910390f35b4561009a565b4361009a565b3a61009a565b4261009a565b4661009a565b61009a6100e7565b4461009a565b6040514181526020016100a4565b60006100f46001436100fa565b40905090565b8181038181111561011b57634e487b7160e01b600052601160045260246000fd5b9291505056fea2646970667358221220a629106cbdbc0017022eedc70f72757902db9dc7881e188747a544aaa638345d64736f6c63430008120033");
        var insufficientGasLimit = BigInteger.valueOf(50000);
        try (var db = new MemoryDatabase(); var statedb = new StateDB(db, Hash.ZERO)) {
            statedb.setBalance(addr1, v5m);
            var result =
                Evm.Apply(statedb, new Invocation(addr1, null, null, input, insufficientGasLimit, false), null);
            assertEquals(
                "unexpected error message",
                "contract creation code storage out of gas",
                result.executionError
            );
            assertEquals("unexpected gas usage", BigInteger.ZERO, result.leftOverGas);
        }
    }
}
