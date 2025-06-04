package io.horizen.evm;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.common.collect.Iterables;
import com.google.common.collect.Iterators;
import io.horizen.evm.params.HandleParams;
import io.horizen.evm.params.OpenStateParams;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.TemporaryFolder;
import org.web3j.rlp.RlpEncoder;
import org.web3j.rlp.RlpString;
import sparkz.crypto.hash.Keccak256;

import java.io.File;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Iterator;

import static org.junit.Assert.*;

public class StateDBTest extends LibEvmTestBase {

    @Rule
    public TemporaryFolder tempFolder = new TemporaryFolder();

    @Test
    public void accountManipulation() throws Exception {
        final var databaseFolder = tempFolder.newFolder("evm-db");

        final var origin = new Address("0xbafe3b6f2a19658df3cb5efca158c93272ff5c0b");

        final var v1234 = BigInteger.valueOf(1234);
        final var v432 = BigInteger.valueOf(432);
        final var v802 = v1234.subtract(v432);
        final var v3 = BigInteger.valueOf(3);
        final var v5 = BigInteger.valueOf(5);

        Hash rootWithBalance1234;
        Hash rootWithBalance802;
        Hash committedRoot;

        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath())) {
            try (var statedb = new StateDB(db, Hash.ZERO)) {
                var intermediateRoot = statedb.getIntermediateRoot();
                assertEquals(
                        "empty state should give the hash of an empty string as the root hash",
                        StateDB.EMPTY_ROOT_HASH,
                        intermediateRoot
                );

                committedRoot = statedb.commit();
                assertEquals("committed root should equal intermediate root", intermediateRoot, committedRoot);
                assertEquals(BigInteger.ZERO, statedb.getBalance(origin));
            }
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                var intermediateRoot = statedb.getIntermediateRoot();
                assertEquals(
                        "empty state should give the hash of an empty string as the root hash",
                        StateDB.EMPTY_ROOT_HASH,
                        intermediateRoot
                );

                committedRoot = statedb.commit();
                assertEquals("committed root should equal intermediate root", intermediateRoot, committedRoot);
                assertEquals(BigInteger.ZERO, statedb.getBalance(origin));
            }
            try (var statedb = new StateDB(db, committedRoot)) {
                statedb.addBalance(origin, v1234);
                assertEquals(v1234, statedb.getBalance(origin));
                assertNotEquals("intermediate root should not equal committed root anymore", committedRoot,
                        statedb.getIntermediateRoot()
                );
                rootWithBalance1234 = statedb.commit();
            }
            try (var statedb = new StateDB(db, rootWithBalance1234)) {
                var revisionId = statedb.snapshot();
                statedb.subBalance(origin, v432);
                assertEquals(v802, statedb.getBalance(origin));
                statedb.revertToSnapshot(revisionId);
                assertEquals(v1234, statedb.getBalance(origin));
                statedb.subBalance(origin, v432);
                assertEquals(v802, statedb.getBalance(origin));

                assertEquals(BigInteger.ZERO, statedb.getNonce(origin));
                statedb.setNonce(origin, v3);
                assertEquals(v3, statedb.getNonce(origin));
                rootWithBalance802 = statedb.commit();
            }
            try (var statedb = new StateDB(db, rootWithBalance802)) {
                statedb.setNonce(origin, v5);
                assertEquals(v5, statedb.getNonce(origin));
            }
            // Verify that automatic resource management worked and StateDB.close() was called.
            // If it was, the handle is invalid now and this should throw.
            assertThrows(
                Exception.class,
                () -> LibEvm.invoke("StateIntermediateRoot", new HandleParams(1), Hash.class).toBytes()
            );
        }
        // also verify that the database was closed
        assertThrows(
            Exception.class,
            () -> LibEvm.invoke("StateOpen", new OpenStateParams(1, Hash.ZERO), int.class)
        );

        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath())) {
            try (var statedb = new StateDB(db, rootWithBalance1234)) {
                assertEquals(v1234, statedb.getBalance(origin));
                assertEquals(BigInteger.ZERO, statedb.getNonce(origin));
            }

            try (var statedb = new StateDB(db, rootWithBalance802)) {
                assertEquals(v802, statedb.getBalance(origin));
                assertEquals(v3, statedb.getNonce(origin));
            }
        }
    }

    @Test
    public void accountStorage() throws Exception {
        final var databaseFolder = tempFolder.newFolder("account-db");
        final var origin = new Address("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff");
        final var key = new Hash("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff010101010101010102020202");
        final Hash[] values = {
            new Hash("0x0000000000000000000000000000000000000000000000000000000000000000"),
            new Hash("0x0000000000000000000000001234000000000000000000000000000000000000"),
            new Hash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
            new Hash("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"),
        };

        Hash initialRoot;
        Hash root;
        var roots = new ArrayList<Hash>();

        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath())) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                assertTrue("account must not exist in an empty state", statedb.isEmpty(origin));
                // make sure the account is not "empty"
                statedb.setNonce(origin, BigInteger.ONE);
                assertFalse("account must exist after nonce increment", statedb.isEmpty(origin));
                initialRoot = statedb.getIntermediateRoot();
                statedb.setStorage(origin, key, values[0]);
                var retrievedValue = statedb.getStorage(origin, key);
                assertEquals(values[0], retrievedValue);
                root = statedb.commit();
                // store the root hash of each state
                roots.add(root);
                var committedValue = statedb.getStorage(origin, key);
                assertEquals(values[0], committedValue);
            }
            for (int i = 1; i < values.length; i++) {
                try (var statedb = new StateDB(db, root)) {
                    var value = values[i];
                    statedb.setStorage(origin, key, value);
                    var retrievedValue = statedb.getStorage(origin, key);
                    assertEquals(value, retrievedValue);
                    root = statedb.commit();
                    // store the root hash of each state
                    roots.add(root);
                    var committedValue = statedb.getStorage(origin, key);
                    assertEquals(value, committedValue);
                }
            }

        }

        // verify that every committed state can be loaded again and that the stored values are still as expected
        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath())) {
            for (int i = 0; i < values.length; i++) {
                try (var statedb = new StateDB(db, roots.get(i))) {
                    var writtenValue = statedb.getStorage(origin, key);
                    assertEquals(values[i], writtenValue);
                    // verify that removing the key results in the initial state root
                    statedb.setStorage(origin, key, null);
                    assertEquals(initialRoot, statedb.getIntermediateRoot());
                }
            }
        }
    }

    @Test
    public void accountStorageEdgeCases() throws Exception {
        final var origin = new Address("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff");
        final var key = new Hash("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff010101010101010102020202");
        // test some negative cases:
        // - trying to store a value that is not 32 bytes should throw - after refactoring to "Hash" this is prevented
        // - writing 32 bytes of zeros and null should behave identical (remove the key-value pair)
        final Hash validValue = new Hash("0x00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff");

        try (var db = new MemoryDatabase()) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                assertTrue("account must not exist in an empty state", statedb.isEmpty(origin));
                // writing to an "empty" account should fail:
                // this is a safety precaution, because empty accounts will be pruned, even if the storage is not empty
                assertThrows(LibEvmException.class, () -> statedb.setStorage(origin, key, validValue));
                // make sure the account is not "empty"
                statedb.setNonce(origin, BigInteger.ONE);
                assertEquals(
                    "reading a non-existent key should return all zeroes",
                    Hash.ZERO,
                    statedb.getStorage(origin, key)
                );
                // make sure this does not throw anymore and the value can be read correctly
                statedb.setStorage(origin, key, validValue);
                assertEquals("value was not written correctly", validValue, statedb.getStorage(origin, key));
                // make sure the value did not change after invalid write attempts
                assertEquals("unexpected change of written value", validValue, statedb.getStorage(origin, key));
                // test removal of the key by using null
                statedb.setStorage(origin, key, null);
                assertEquals("value should be all zeroes", Hash.ZERO, statedb.getStorage(origin, key));
                // write the value again
                statedb.setStorage(origin, key, validValue);
                // test removal of the key by using all zeros
                statedb.setStorage(origin, key, Hash.ZERO);
                assertEquals("value should be all zeroes", Hash.ZERO, statedb.getStorage(origin, key));
            }
        }
    }

    private void testAccessListAccounts(StateDB statedb, Address sender, Address destination, Address other, Address coinbase, boolean isShanghai) {
        final var key1 = new Hash("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff000000000000000000000001");
        final var key2 = new Hash("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff000000000000000000000002");

        statedb.accessSetup(sender, destination, coinbase, new ForkRules(isShanghai));
        assertTrue("sender must be on access list", statedb.accessAccount(sender));
        assertTrue("destination must be on access list", statedb.accessAccount(destination));
        if (isShanghai){
            assertTrue("coinbase must be on access list when Shanghai is active", statedb.accessAccount(coinbase));
        }
        else {
            assertFalse("coinbase must not be on access list when Shanghai is not active", statedb.accessAccount(coinbase));
        }
        assertFalse(
            "sender storage slot must not be on access list before first access",
            statedb.accessSlot(sender, key1)
        );
        assertTrue(
            "sender storage slot must be on access list after first access",
            statedb.accessSlot(sender, key1)
        );
        assertFalse(
            "sender storage slot must not be on access list before first access",
            statedb.accessSlot(sender, key2)
        );
        assertTrue(
            "sender storage slot must be on access list after first access",
            statedb.accessSlot(sender, key2)
        );

        assertFalse(
            "other account must not be on access list before first access",
            statedb.accessAccount(other)
        );
        assertTrue(
            "other account must be on access list after first access",
            statedb.accessAccount(other)
        );
        assertFalse(
            "other storage slot must not be on access list before first access",
            statedb.accessSlot(other, key1)
        );
        assertTrue(
            "other storage slot must be on access list after first access",
            statedb.accessSlot(other, key1)
        );
    }

    @Test
    public void accessList() throws Exception {
        final var accounts = new Address[] {
            new Address("0x0011001100110011001100110011001100110011"),
            new Address("0x0022002200220022002200220022002200220022"),
            new Address("0x0033003300330033003300330033003300330033"),
        };
        Address coinbase = new Address("0xcc110011001100110011001100110011001100aa");
        boolean isShanghai = false;

        try (var db = new MemoryDatabase()) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                // test multiple permutations of the accounts in a row to make sure the access list is correctly reset
                testAccessListAccounts(statedb, accounts[0], accounts[1], accounts[2], coinbase, isShanghai);
                testAccessListAccounts(statedb, accounts[1], accounts[2], accounts[0], coinbase, isShanghai);
                testAccessListAccounts(statedb, accounts[2], accounts[0], accounts[1], coinbase, isShanghai);
                testAccessListAccounts(statedb, accounts[0], accounts[2], accounts[1], coinbase, isShanghai);
                testAccessListAccounts(statedb, accounts[1], accounts[0], accounts[2], coinbase, isShanghai);
            }
        }

        isShanghai = true;
        try (var db = new MemoryDatabase()) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                // test multiple permutations of the accounts in a row to make sure the access list is correctly reset
                testAccessListAccounts(statedb, accounts[0], accounts[1], accounts[2], coinbase, isShanghai);
                testAccessListAccounts(statedb, accounts[1], accounts[2], coinbase, accounts[0], isShanghai);
                testAccessListAccounts(statedb, accounts[2], coinbase, accounts[0], accounts[1], isShanghai);
                testAccessListAccounts(statedb, coinbase, accounts[0], accounts[1], accounts[2], isShanghai);
                testAccessListAccounts(statedb, accounts[0], accounts[2], accounts[1], coinbase, isShanghai);
                testAccessListAccounts(statedb, accounts[1], accounts[0], accounts[2], coinbase, isShanghai);
                testAccessListAccounts(statedb, coinbase, accounts[1], accounts[0], accounts[2], isShanghai);
            }
        }

    }

    @Test
    public void testAccountTypes() throws Exception {
        final var code = bytes("aa87aee0394326416058ef46b907882903f3646ef2a6d0d20f9e705b87c58c77");
        final var addr1 = new Address("0x1234561234561234561234561234561234561230");

        try (var db = new MemoryDatabase()) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                // Test 1: non-existing account is an EOA account
                assertTrue("EOA account expected", statedb.isEoaAccount(addr1));
                assertFalse("EOA account expected", statedb.isSmartContractAccount(addr1));

                // Test 2: account exists and has NO code defined, so considered as EOA
                // Declare account with some coins
                statedb.addBalance(addr1, BigInteger.TEN);
                assertTrue("EOA account expected", statedb.isEoaAccount(addr1));
                assertFalse("EOA account expected", statedb.isSmartContractAccount(addr1));

                // Test 3: Account exists and has code defined, so considered as Smart contract account
                statedb.setCode(addr1, code);
                assertFalse("Smart contract account expected", statedb.isEoaAccount(addr1));
                assertTrue("Smart contract account expected", statedb.isSmartContractAccount(addr1));
            }
        }
    }

    @Test
    public void proof() throws Exception {
        final var address = new Address("0xcca577ee56d30a444c73f8fc8d5ce34ed1c7da8b");

        try (var db = new MemoryDatabase()) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                statedb.addBalance(address, BigInteger.TEN);
                statedb.setStorage(
                    address,
                    Hash.ZERO,
                    padToHash(RlpEncoder.encode(RlpString.create(bytes("94de74da73d5102a796559933296c73e7d1c6f37fb"))))
                );
                statedb.setStorage(
                    address,
                    new Hash((byte[]) Keccak256.hash(bytes(
                        "0000000000000000000000000000000000000000000000000000000000000001"))),
                    padToHash(RlpEncoder.encode(RlpString.create(bytes("02"))))
                );

               var root =  statedb.commit();

                // this should return the account proof with code hash, updated balance, empty storageProof
                var resultA = statedb.getProof(address, root,null);
                assertEquals(BigInteger.TEN, resultA.balance);
                assertEquals(1, resultA.accountProof.length);
                assertEquals(0, resultA.storageProof.length);

                var resultB = statedb.getProof(address, root, new Hash[] {Hash.ZERO});
                assertEquals(BigInteger.TEN, resultB.balance);
                assertEquals(1, resultB.accountProof.length);
                assertEquals(1, resultB.storageProof.length);
                assertEquals(2, resultB.storageProof[0].proof.length);
                assertEquals(
                    "0xf8518080a0cd5ca8a057cdc46dd649378e1cfbf1e166d25321c859d2918c955c48baf18f2d8080808080808080a08b7da99e493f28ab906f59591750bb1f2058beb82ab748f1cc2b66d4cf499f608080808080",
                    resultB.storageProof[0].proof[0]
                );
                assertEquals(
                    "0xf839a0390decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56397969594de74da73d5102a796559933296c73e7d1c6f37fb",
                    resultB.storageProof[0].proof[1]
                );
            }
        }
    }

    @Test
    public void dump() throws Exception {
        final var databaseFolder = tempFolder.newFolder("account-db");
        final var key = new Hash("0xbafe3b6f2a19658df3cb5efca158c93272ff5cff010101010101010102020202");
        final Hash[] values = {
                new Hash("0x0000000000000000000000000000000000000000000000000000000000000001"),
        };
        final var codeString = "aa87aee0394326416058ef46b907882903f3646ef2a6d0d20f9e705b87c58c77";
        final var code = bytes(codeString);

        Hash rootHash;
        ArrayList<Address> addresses = new ArrayList<>();
        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath(), true)) {
            try (var statedb = new StateDB(db, StateDB.EMPTY_ROOT_HASH)) {
                for (int i = 10; i < 20; i++) {
                    var address = new Address("0xbafe3b6f2a19658df3cb5efca158c93272ff5c" + i);
                    addresses.add(address);
                    statedb.setNonce(address, BigInteger.ONE);
                    statedb.setCode(address, code);
                    for (int j = 0; j < values.length; j++) {
                        var value = values[j];
                        statedb.setStorage(address, key, value);
                    }
                }
                rootHash = statedb.commit();
                // store the root hash of each state
            }

        }

        String dumpFile = System.getProperty("java.io.tmpdir") + File.separator + "dump.json";
        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath(), true)) {
            try (var statedb = new StateDB(db, rootHash)) {
                statedb.dump(dumpFile);
            }
        }

        byte[] jsonData = Files.readAllBytes(Paths.get(dumpFile));
        ObjectMapper objectMapper = new ObjectMapper();
        JsonNode jsonTreeRootNode = objectMapper.readTree(jsonData);
        JsonNode rootFieldNode = jsonTreeRootNode.path("root");
        assertEquals(rootHash.toStringNoPrefix(), rootFieldNode.asText());

        JsonNode accountsNode = jsonTreeRootNode.path("accounts");
        Iterator<JsonNode> elements = accountsNode.elements();
        assertEquals(addresses.size(), Iterators.size(elements));
        for (Address a: addresses) {
            JsonNode addressNode = accountsNode.get(a.toString());
            assertEquals(0, addressNode.get("balance").asInt());
            assertEquals(1, addressNode.get("nonce").asInt());
            assertEquals("0x" + codeString, addressNode.get("code").asText());
            assertEquals(values.length, Iterators.size(addressNode.get("storage").elements()));
        }

        // test that if the file is not writable, executing dump results in an error

        File readonlyFile = tempFolder.newFile("readonly.json");
        readonlyFile.setReadOnly();
        try (var db = new LevelDBDatabase(databaseFolder.getAbsolutePath(), true)) {
            try (var statedb = new StateDB(db, rootHash)) {
                assertThrows(
                        LibEvmException.class,
                        () -> statedb.dump(readonlyFile.getAbsolutePath())
                ).getMessage().contains("access denied");
            }
        }
    }


}
