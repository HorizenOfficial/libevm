package lib

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/params"
	"math"
	"math/big"
	"time"
)

type EvmParams struct {
	HandleParams
	From         common.Address  `json:"from"`
	To           *common.Address `json:"to"`
	Value        *hexutil.Big    `json:"value"`
	Input        []byte          `json:"input"`
	AvailableGas hexutil.Uint64  `json:"availableGas"`
	GasPrice     *hexutil.Big    `json:"gasPrice"`
	Context      EvmContext      `json:"context"`
}

type BlockHashCallback struct{ Callback }

func (c *BlockHashCallback) getBlockHash(blockNumber uint64) common.Hash {
	blockNumberBig := new(big.Int).SetUint64(blockNumber)
	if c == nil {
		// fallback to mocked block hash
		return common.BytesToHash(crypto.Keccak256([]byte(blockNumberBig.String())))
	}
	blockNumberHex := (*hexutil.Big)(blockNumberBig).String()
	return common.HexToHash(c.Invoke(blockNumberHex))
}

type EvmContext struct {
	ChainID           hexutil.Uint64     `json:"chainID"`
	Coinbase          common.Address     `json:"coinbase"`
	GasLimit          hexutil.Uint64     `json:"gasLimit"`
	BlockNumber       *hexutil.Big       `json:"blockNumber"`
	Time              *hexutil.Big       `json:"time"`
	BaseFee           *hexutil.Big       `json:"baseFee"`
	Random            common.Hash        `json:"random"`
	BlockHashCallback *BlockHashCallback `json:"blockHashCallback"`
	Tracer            *int               `json:"tracer"`
}

// setDefaults for parameters that were omitted
func (p *EvmParams) setDefaults() {
	if p.Value == nil {
		p.Value = (*hexutil.Big)(common.Big0)
	}
	if p.AvailableGas == 0 {
		p.AvailableGas = (hexutil.Uint64)(math.MaxInt64)
	}
	if p.GasPrice == nil {
		p.GasPrice = (*hexutil.Big)(common.Big0)
	}
	p.Context.setDefaults()
}

// setDefaults for parameters that were omitted
func (c *EvmContext) setDefaults() {
	if c.GasLimit == 0 {
		c.GasLimit = (hexutil.Uint64)(math.MaxInt64)
	}
	if c.BlockNumber == nil {
		c.BlockNumber = (*hexutil.Big)(common.Big0)
	}
	if c.Time == nil {
		c.Time = (*hexutil.Big)(big.NewInt(time.Now().Unix()))
	}
	if c.BaseFee == nil {
		c.BaseFee = (*hexutil.Big)(big.NewInt(params.InitialBaseFee))
	}
}

func (c *EvmContext) getBlockContext() vm.BlockContext {
	return vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     c.BlockHashCallback.getBlockHash,
		Coinbase:    c.Coinbase,
		GasLimit:    uint64(c.GasLimit),
		BlockNumber: c.BlockNumber.ToInt(),
		Time:        c.Time.ToInt(),
		Difficulty:  common.Big0,
		BaseFee:     c.BaseFee.ToInt(),
		Random:      &c.Random,
	}
}

func (c *EvmContext) getChainConfig() *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:             new(big.Int).SetUint64(uint64(c.ChainID)),
		HomesteadBlock:      common.Big0,
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         common.Big0,
		EIP155Block:         common.Big0,
		EIP158Block:         common.Big0,
		ByzantiumBlock:      common.Big0,
		ConstantinopleBlock: common.Big0,
		PetersburgBlock:     common.Big0,
		IstanbulBlock:       common.Big0,
		MuirGlacierBlock:    common.Big0,
		BerlinBlock:         common.Big0,
		LondonBlock:         common.Big0,
	}
}

// getTracer retrieves an instance of a tracer if a handle is given
func (c *EvmContext) getTracer(s *Service) (tracers.Tracer, error) {
	if c.Tracer == nil {
		return nil, nil
	}
	err, tracerPtr := s.tracers.Get(*c.Tracer)
	if err != nil {
		return nil, err
	}
	return *tracerPtr, nil
}

type EvmResult struct {
	UsedGas         hexutil.Uint64  `json:"usedGas"`
	EvmError        string          `json:"evmError"`
	ReturnData      []byte          `json:"returnData"`
	ContractAddress *common.Address `json:"contractAddress"`
	Reverted        bool            `json:"reverted"`
}

func (s *Service) EvmApply(params EvmParams) (error, *EvmResult) {
	// apply defaults to missing parameters
	params.setDefaults()

	err, statedb := s.statedbs.Get(params.Handle)
	if err != nil {
		return err, nil
	}

	// get tracer if a handle is given
	tracer, err := params.Context.getTracer(s)
	if err != nil {
		return err, nil
	}

	var (
		txContext = vm.TxContext{
			Origin:   params.From,
			GasPrice: params.GasPrice.ToInt(),
		}
		blockContext = params.Context.getBlockContext()
		chainConfig  = params.Context.getChainConfig()
		evmConfig    = vm.Config{
			Debug:                   tracer != nil,
			Tracer:                  tracer,
			NoBaseFee:               false,
			EnablePreimageRecording: false,
			JumpTable:               nil,
			ExtraEips:               nil,
			ExternalContracts:       []common.Address{common.HexToAddress("0x0000000000000000000022222222222222222222")},
			ExternalCallback: func(caller common.Address, callee common.Address, value *big.Int, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
				mockedForgerStakesData := common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000bc22971e6a19a3ddf28dafcbe3eaf261cfac0f3dc07f9cef79dfc94175d1eb8cc000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000039ef4eea2b229c11dba18392b7f12c83c89053b9ffc4065704d28337afdb9e216e264be8e1121f29b33a84860b78008cf8638ebe0bfce2073aa51bc5854f8f297e31b89358a29e27f95380b91268e68dc58d1d0a8000000000000000000000000000000000000000000000000000000000000000fe9436b7f4645cc5562e7f37996a11a63da703043df985ec23f9c5a642a288ff000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000002a789f245142753075c4c5a3c603b52d8f01d361b85bb1e0b3de12c8bb3b7f7cb04ee5f7cdec4de4bf0879536824cc84adf6d7d29e9f78162e4509a0df40e3d8525801c363a8d327ce0c9eb4da8a3615489aee1680000000000000000000000000000000000000000000000000000000000000000e19a05f03af33541c3219a734b0868f39b36493b3e8ba63c86c5659725d582c000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000551c0294ef40e7d4da0d0e62f77141cdb455dcc47b48711bc08c786c486520bb9eb69ba43ed04fc2cc617c24e85d757ffcaf9a971644ba05f4fcea365fabba5b1ee38f0a1db512f7178c3733e8680589cb4d7f350000000000000000000000000000000000000000000000000000000000000000aa27870083d34abec0abf7a0a5e39d2cb353ffb96c486af53868783147211f18000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000072b6b2ed1c4d951a5247206b771a2615ec25a255087a518986785e8746d114d6cbdea0fc9f048362b1b318b4d36f7d1d51596ac789c3617a58b703631dcff9ce96697e0b4e1c7ce9207db22edf110e5b29347c1580000000000000000000000000000000000000000000000000000000000000001ddda8086cbc1d0752e95651ba810b7a56441bf42a770010475f8d02774f7411000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000e36c4e40d357e4a6e4df7a05f5944264df867265c0905c48964a940880e8de76c84dddc8535d0878a693097b3e86f6509e649606b5c5916187e905d83ba75ebdc39d2d526d6402a4801677f325dd55e1df5d810d000000000000000000000000000000000000000000000000000000000000000086c07dadb86c083eb56e6e9ac3891afab10806b8221b50a5763ad56748b51113000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000002e0126cbcae91e490522f65925773d182339799674cac9966adc776d28720f7bfa744e7c725b7388734bb6801b98f3470aefdf3f0ce9e0bd37fb0abcc3dc392baeeb8ac3acd47ac73ae79429cd9f4807faec06330000000000000000000000000000000000000000000000000000000000000000fcbd1f41cf3f859f0f3c3de58d4227032400e9edc46feb81ad0d9790283be923000000000000000000000000000000000000000000000000016345785d8a00000000000000000000000000008bf4857b740943b005fbbee80719421ab5b3dc510ef5b3566378c969af6545b06706e380cba6ba6dc55d7c1a5ae0c914475100c7c7b855d86e93a45d0108026614c4a8e2acb201f7c4e0f26297e3fbde0cf57e268000000000000000000000000000000000000000000000000000000000000000703bf3e9519b03c8e97f29d5e4cce7e136f894a6fd5eb35020f193ffb77b838c000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000008a698b7535d6eeb6341e5c65e220c49f177f73e6cd00ba09b5d2b771c1f90399e183bf80e54886a390904be2e018d0d80b46b018d71700c48baf671f9ebd48895f555f32b7b0c20ee2a3dcb35290c9a412a2e03800000000000000000000000000000000000000000000000000000000000000006f64a8254e75c9dd9dba6499112c0fe99b7ad5f1d978ace402c3459d1d09891000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000030d617085728a81563dfd655c35020ab8b78ecec76f4fb6b460dbc0fdf3e17fb5b9888ab9ce2f2f009e2fb32ef95eff196b4326eacde2c72c65e185d9e04e6012ce0bc6165a17a4f95215d701982c8b8cb5d970800000000000000000000000000000000000000000000000000000000000000008d546419e013bafbaead980365902a624f724cd1611086ab55946ec72cbfe889000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000e089abfcb53895adcbc8825aba391ebd1ddd83ed12e73f518bd4314b5251d881394711433f348ebea66b817aade3c93ffb3650dc51c95d2cd6428da7189c3e7451e51906bde78fee932f3a61644c22ca8856f0320000000000000000000000000000000000000000000000000000000000000000fda9653e35779b6b8f5186548d9bd3e7f66f6bc662d530569104cc1ad6afe437000000000000000000000000000000000000000000000000016345785d8a000000000000000000000000000075413637bc6e1818b36c9acf56fac28765bc3ec8fa9471609398bc7520e8151a96409151433d426a094fc70e6f32b89ed854d12acd7a2250e51d6ef985275847e7c715e67b4967a44910d19950c19bf43578a3210000000000000000000000000000000000000000000000000000000000000000")
				return mockedForgerStakesData, gas, nil
			},
		}
		evm              = vm.NewEVM(blockContext, txContext, statedb, chainConfig, evmConfig)
		sender           = vm.AccountRef(params.From)
		gas              = uint64(params.AvailableGas)
		contractCreation = params.To == nil
	)

	var (
		returnData      []byte
		vmerr           error
		contractAddress *common.Address
	)
	if contractCreation {
		// The following nonce modification is a workaround for the following problem:
		//
		// Creating a smart contract should increment the callers' nonce, this is true for EOAs as well as contracts
		// creating other contracts. Thus, the nonce increment is done in evm.Create and must be there.
		// In contrast to that behavior, for the top level call the nonce was already increased by the SDK at this
		// point. So if we don't do anything here the nonce of an EOA will be increased twice when a smart contract is
		// deployed.
		//
		// As the contract address is calculated from the nonce we can't just decrement the nonce afterwards (to undo
		// the unwanted change), we have to do that before running the EVM. This also introduces two edge cases:
		//
		// - The check nonce > 0 was necessary in an earlier version where the nonce was NOT increased when the call was
		//   performed in the context of eth_call via RPC. This is fixed now, but we should still keep this as a
		//   precaution (this would cause unsigned integer underflow to maxUint64) and because it is useful for tests.
		//
		// - The EVM.create call can fail before it even reaches the point of incrementing the nonce. We have to make
		//   sure to NOT decrement the nonce in that case. Hence, setting the nonce to the value before the EVM call in
		//   case it was modified.
		nonce := statedb.GetNonce(params.From)
		if nonce > 0 {
			statedb.SetNonce(params.From, nonce-1)
		}
		// we ignore returnData here because it holds the contract code that was just deployed
		var deployedContractAddress common.Address
		_, deployedContractAddress, gas, vmerr = evm.Create(sender, params.Input, gas, params.Value.ToInt())
		contractAddress = &deployedContractAddress
		// if there is an error evm.Create might not have incremented the nonce as expected,
		if statedb.GetNonce(params.From) != nonce {
			statedb.SetNonce(params.From, nonce)
		}
	} else {
		returnData, gas, vmerr = evm.Call(sender, *params.To, params.Input, gas, params.Value.ToInt())
	}

	// no error means successful transaction, otherwise failure
	evmError := ""
	if vmerr != nil {
		evmError = vmerr.Error()
	}

	if returnData == nil {
		// we want [] rather than null in the response
		returnData = make([]byte, 0)
	}

	return nil, &EvmResult{
		UsedGas:         params.AvailableGas - hexutil.Uint64(gas),
		EvmError:        evmError,
		ReturnData:      returnData,
		ContractAddress: contractAddress,
		Reverted:        vmerr == vm.ErrExecutionReverted,
	}
}
