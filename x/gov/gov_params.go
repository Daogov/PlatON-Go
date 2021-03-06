// Copyright 2018-2019 The PlatON Network Authors
// This file is part of the PlatON-Go library.
//
// The PlatON-Go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The PlatON-Go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the PlatON-Go library. If not, see <http://www.gnu.org/licenses/>.

package gov

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/PlatONnetwork/PlatON-Go/params"

	"github.com/PlatONnetwork/PlatON-Go/log"

	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/core/snapshotdb"
	"github.com/PlatONnetwork/PlatON-Go/x/xcom"
)

var (
	initGovParam sync.Once
)

var governParam []*GovernParam

func queryInitParam() []*GovernParam {
	initGovParam.Do(func() {
		log.Info("Init Govern parameters ...")
		governParam = initParam()
	})
	return governParam
}

func initParam() []*GovernParam {
	return []*GovernParam{

		/**
		About Staking module
		*/
		{

			ParamItem: &ParamItem{ModuleStaking, KeyStakeThreshold,
				fmt.Sprintf("minimum amount of stake, range：[%d, %d) ", xcom.MillionLAT, xcom.TenMillionLAT)},
			ParamValue: &ParamValue{"", xcom.StakeThreshold().String(), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				threshold, ok := new(big.Int).SetString(value, 10)
				if !ok {
					return fmt.Errorf("Parsed StakeThreshold is failed")
				}

				if err := xcom.CheckStakeThreshold(threshold); nil != err {
					return err
				}
				return nil
			},
		},

		{
			ParamItem: &ParamItem{ModuleStaking, KeyOperatingThreshold,
				fmt.Sprintf("minimum amount of stake increasing funds, delegation funds, or delegation withdrawing funds, range：[%d, %d) ", xcom.TenLAT, xcom.TenThousandLAT)},
			ParamValue: &ParamValue{"", xcom.OperatingThreshold().String(), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				threshold, ok := new(big.Int).SetString(value, 10)
				if !ok {
					return fmt.Errorf("Parsed OperatingThreshold is failed")
				}

				if err := xcom.CheckOperatingThreshold(threshold); nil != err {
					return err
				}

				return nil

			},
		},

		{
			ParamItem: &ParamItem{ModuleStaking, KeyMaxValidators,
				fmt.Sprintf("maximum amount of validator, range：[%d, %d]", xcom.MaxConsensusVals(), xcom.CeilMaxValidators)},
			ParamValue: &ParamValue{"", strconv.Itoa(int(xcom.MaxValidators())), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				num, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed MaxValidators is failed: %v", err)
				}

				if err := xcom.CheckMaxValidators(num); nil != err {
					return err
				}

				return nil

			},
		},

		{
			ParamItem: &ParamItem{ModuleStaking, KeyUnStakeFreezeDuration,
				fmt.Sprintf("quantity of epoch for skake withdrawal, range：(MaxEvidenceAge, %d]", xcom.CeilUnStakeFreezeDuration)},
			ParamValue: &ParamValue{"", strconv.Itoa(int(xcom.UnStakeFreezeDuration())), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				num, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed UnStakeFreezeDuration is failed: %v", err)
				}

				age, _ := GovernMaxEvidenceAge(blockNumber, blockHash)

				if err := xcom.CheckUnStakeFreezeDuration(num, int(age)); nil != err {
					return err
				}

				return nil

			},
		},

		/**
		About Slashing module
		*/

		{
			ParamItem: &ParamItem{ModuleSlashing, KeySlashFractionDuplicateSign,
				fmt.Sprintf("quantity of base point(1BP=1‱). Node's stake will be deducted(BPs*staking amount*1‱) it the node sign block duplicatlly, range：(%d, %d]", xcom.Zero, xcom.TenThousand)},
			ParamValue: &ParamValue{"", strconv.Itoa(int(xcom.SlashFractionDuplicateSign())), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				fraction, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed SlashFractionDuplicateSign is failed: %v", err)
				}

				if err := xcom.CheckSlashFractionDuplicateSign(fraction); nil != err {
					return err
				}

				return nil
			},
		},

		{
			ParamItem: &ParamItem{ModuleSlashing, KeyDuplicateSignReportReward,
				fmt.Sprintf("quantity of base point(1bp=1%%). Bonus(BPs*deduction amount for sign block duplicatlly*%%) to the node who reported another's duplicated-signature, range：(%d, %d]", xcom.Zero, xcom.Eighty)},
			ParamValue: &ParamValue{"", strconv.Itoa(int(xcom.DuplicateSignReportReward())), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				fraction, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed DuplicateSignReportReward is failed: %v", err)
				}

				if err := xcom.CheckDuplicateSignReportReward(fraction); nil != err {
					return err
				}

				return nil
			},
		},

		{
			ParamItem: &ParamItem{ModuleSlashing, KeyMaxEvidenceAge,
				fmt.Sprintf("quantity of epoch. During these epochs after a node duplicated-sign, others can report it, range：(%d, UnStakeFreezeDuration)", xcom.Zero)},
			ParamValue: &ParamValue{"", strconv.Itoa(int(xcom.MaxEvidenceAge())), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				age, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed MaxEvidenceAge is failed: %v", err)
				}

				duration, _ := GovernUnStakeFreezeDuration(blockNumber, blockHash)

				if err := xcom.CheckMaxEvidenceAge(age, int(duration)); nil != err {
					return err
				}

				return nil

			},
		},
		{
			ParamItem: &ParamItem{ModuleSlashing, KeySlashBlocksReward,
				fmt.Sprintf("quantity of block, the total bonus amount for these blocks will be deducted from a inefficient node's stake, range：[%d, %d)", xcom.Zero, xcom.CeilBlocksReward)},
			ParamValue: &ParamValue{"", strconv.Itoa(int(xcom.SlashBlocksReward())), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				rewards, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed SlashBlocksReward is failed: %v", err)
				}

				if err := xcom.CheckSlashBlocksReward(rewards); nil != err {
					return err
				}

				return nil

			},
		},

		/**
		About Block module
		*/
		{
			ParamItem:  &ParamItem{ModuleBlock, KeyMaxBlockGasLimit, fmt.Sprintf("maximum gas limit per block, range：[%d, %d]", int(params.GenesisGasLimit), int(params.MaxGasCeil))},
			ParamValue: &ParamValue{"", strconv.Itoa(int(params.DefaultMinerGasCeil)), 0},
			ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {

				gasLimit, err := strconv.Atoi(value)
				if nil != err {
					return fmt.Errorf("Parsed MaxBlockGasLimit is failed: %v", err)
				}

				// (4712388<= x < =21000 0000)
				if gasLimit < int(params.GenesisGasLimit) || gasLimit > int(params.MaxGasCeil) {
					return common.InvalidParameter.Wrap(fmt.Sprintf("The MaxBlockGasLimit must be [%d, %d]", int(params.GenesisGasLimit), int(params.MaxGasCeil)))
				}

				return nil
			},
		},

		/**
		About TxPool module
		*/
		//{
		//	ParamItem:  &ParamItem{ModuleTxPool, KeyMaxTxDataLimit, fmt.Sprintf("maximum data length per transaction, range：(%d, %d]", xcom.Zero, CeilTxSize)},
		//	ParamValue: &ParamValue{"", strconv.Itoa(GenesisTxSize), 0},
		//	ParamVerifier: func(blockNumber uint64, blockHash common.Hash, value string) error {
		//
		//		txSize, err := strconv.Atoi(value)
		//		if nil != err {
		//			return fmt.Errorf("Parsed MaxTxDataLimit is failed: %v", err)
		//		}
		//
		//		// (0, 10MB]
		//		if txSize > CeilTxSize {
		//			return common.InvalidParameter.Wrap(fmt.Sprintf("The MaxTxDataLimit must be (%d, %d]", xcom.Zero, CeilTxSize))
		//		}
		//
		//		return nil
		//	},
		//},
	}
}

var ParamVerifierMap = make(map[string]ParamVerifier)

func InitGenesisGovernParam(snapDB snapshotdb.DB) error {
	var paramItemList []*ParamItem
	for _, param := range queryInitParam() {
		paramItemList = append(paramItemList, param.ParamItem)

		key := KeyParamValue(param.ParamItem.Module, param.ParamItem.Name)
		value := common.MustRlpEncode(param.ParamValue)
		if err := snapDB.PutBaseDB(key, value); err != nil {
			return err
		}
	}

	key := KeyParamItems()
	value := common.MustRlpEncode(paramItemList)
	if err := snapDB.PutBaseDB(key, value); err != nil {
		return err
	}
	return nil
}

func RegisterGovernParamVerifiers() {
	for _, param := range queryInitParam() {
		RegGovernParamVerifier(param.ParamItem.Module, param.ParamItem.Name, param.ParamVerifier)
	}
}

func RegGovernParamVerifier(module, name string, callback ParamVerifier) {
	ParamVerifierMap[module+"/"+name] = callback
}
