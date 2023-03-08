# Changelog

## 0.1.0

Provides standlone access to go-ethereum features:
- `StateDB` with underlying database:
  - `MemoryDatabase`
  - `LevelDBDatabase`
- `Evm`
  - operates on `StateDB`
  - gas consumption is tracked and returned
  - **not** handled and left to the caller:
    - most validation of inputs
    - intrinsic gas
    - transfer of fees
- `TrieHasher`
  - builds an ad-hoc merkle trie and returns the root hash

Based on go-ethereum `v1.10.26`

Also provides vital types for the Ethereum ecosystem:
- `Address`
  - immutable 20 bytes
- `Hash`
  - immutable 32 bytes
- Ethereum RPC API compatible Jackson JSON (de-)serializers for:
  - `Address`
    - `0x0123456789012345678901234567890123456789`
  - `Hash`
    - `0x0123456701234567012345670123456701234567012345670123456701234567`
  - `BigInteger` aka. "Quantity"
    - `0x0` ... `0x1234`
