package codec

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ptypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashtype "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtype "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplyexp "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type (
	PubKey         = crypto.PubKey
	PrivKey        = crypto.PrivKey
	Msg            = sdk.Msg
	Account        = auth.Account
	VestingAccount = auth.VestingAccount
	Content        = govtypes.Content

	DuplicateVoteEvidence   = tmtypes.DuplicateVoteEvidence
	PrivKeyEd25519          = ed25519.PrivKeyEd25519
	PrivKeySecp256k1        = secp256k1.PrivKeySecp256k1
	PubKeyEd25519           = ed25519.PubKeyEd25519
	PubKeySecp256k1         = secp256k1.PubKeySecp256k1
	PubKeyMultisigThreshold = multisig.PubKeyMultisigThreshold
	SignedMsgType           = tmtypes.SignedMsgType
	VoteOption              = govtypes.VoteOption
	Vote                    = tmtypes.Vote

	SdkInt = sdk.Int
	SdkDec = sdk.Dec

	Tx             = sdk.Tx
	SupplyI        = supplyexp.SupplyI
	ModuleAccountI = supplyexp.ModuleAccountI

	ConsAddress  = sdk.ConsAddress
	Coin         = sdk.Coin
	DecCoin      = sdk.DecCoin
	StdSignature = auth.StdSignature
	ParamChange  = ptypes.ParamChange
	Input        = bank.Input
	Output       = bank.Output
	AccAddress   = sdk.AccAddress

	BaseAccount                    = auth.BaseAccount
	BaseVestingAccount             = auth.BaseVestingAccount
	ContinuousVestingAccount       = auth.ContinuousVestingAccount
	DelayedVestingAccount          = auth.DelayedVestingAccount
	StdTx                          = auth.StdTx
	MsgBeginRedelegate             = staking.MsgBeginRedelegate
	MsgCreateValidator             = staking.MsgCreateValidator
	MsgDelegate                    = staking.MsgDelegate
	MsgEditValidator               = staking.MsgEditValidator
	MsgUndelegate                  = staking.MsgUndelegate
	MsgUnjail                      = slashing.MsgUnjail
	MsgSetWithdrawAddress          = distr.MsgSetWithdrawAddress
	MsgWithdrawDelegatorReward     = distr.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorCommission = distr.MsgWithdrawValidatorCommission
	MsgDeposit                     = gov.MsgDeposit
	MsgSubmitProposal              = gov.MsgSubmitProposal
	MsgVote                        = gov.MsgVote
	SoftwareUpgradeProposal        = gov.SoftwareUpgradeProposal
	TextProposal                   = gov.TextProposal
	ParameterChangeProposal        = ptypes.ParameterChangeProposal
	CommunityPoolSpendProposal     = distr.CommunityPoolSpendProposal
	FeePool                        = distr.FeePool
	MsgMultiSend                   = bank.MsgMultiSend
	MsgSend                        = bank.MsgSend
	MsgVerifyInvariant             = crisis.MsgVerifyInvariant
	Supply                         = supply.Supply
	ModuleAccount                  = supply.ModuleAccount

	AccAddressList = []AccAddress
	CommitInfo     = rootmulti.CommitInfo
	StoreInfo      = rootmulti.StoreInfo

	Validator                  = stakingtype.Validator
	Delegation                 = stakingtype.Delegation
	BondStatus                 = sdk.BondStatus
	DelegatorStartingInfo      = distr.DelegatorStartingInfo
	ValidatorHistoricalRewards = distr.ValidatorHistoricalRewards
	ValidatorCurrentRewards    = distr.ValidatorCurrentRewards
	ValidatorSlashEvent        = distr.ValidatorSlashEvent
	DecCoins                   = sdk.DecCoins
	ValidatorSigningInfo       = slashtype.ValidatorSigningInfo
	ValAddress                 = sdk.ValAddress
	ValAddressList             = []ValAddress
)
