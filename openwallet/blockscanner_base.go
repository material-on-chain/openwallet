/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwallet

import (
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/timer"
	"sync"
	"time"
)

const (
	periodOfTask = 5 * time.Second //定时任务执行隔间
)

//BlockScannerBase 区块链扫描器基本结构实现
type BlockScannerBase struct {
	AddressInScanning map[string]string                    //加入扫描的地址
	scanTask          *timer.TaskTimer                     //扫描定时器
	Mu                sync.RWMutex                         //读写锁
	Observers         map[BlockScanNotificationObject]bool //观察者
	Scanning          bool                                 //是否扫描中
	PeriodOfTask      time.Duration
}

//ExtractResult 扫描完成的提取结果
type ExtractResult struct {
	Recharges   []*Recharge
	TxID        string
	BlockHeight uint64
	Success     bool
	Reason      string
}

//SaveResult 保存结果
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

//NewBTCBlockScanner 创建区块链扫描器
func NewBlockScannerBase() *BlockScannerBase {
	bs := BlockScannerBase{}
	bs.AddressInScanning = make(map[string]string)
	//bs.WalletInScanning = make(map[string]*WalletWrapper)
	bs.Observers = make(map[BlockScanNotificationObject]bool)
	bs.PeriodOfTask = periodOfTask
	return &bs
}

//AddAddress 添加订阅地址
func (bs *BlockScannerBase) AddAddress(address, sourceKey string) error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	bs.AddressInScanning[address] = sourceKey
	return nil
	//if _, exist := bs.WalletInScanning[sourceKey]; exist {
	//	return
	//}
	//bs.WalletInScanning[sourceKey] = wrapper
}

//AddWallet 添加扫描钱包
//func (bs *BlockScannerBase) AddWallet(sourceKey string, wrapper *WalletWrapper) {
//	bs.Mu.Lock()
//	defer bs.Mu.Unlock()
//
//	if _, exist := bs.WalletInScanning[sourceKey]; exist {
//		//已存在，不重复订阅
//		return
//	}
//
//	bs.WalletInScanning[sourceKey] = wrapper
//
//	//删除充值记录
//	//wallet.DropRecharge()
//
//	//导入钱包该账户的所有地址
//	addrs, err := wrapper.GetAddressList(0, -1)
//	if err != nil {
//		return
//	}
//
//	log.Std.Info("block scanner load wallet [%s] existing addresses: %d ", sourceKey, len(addrs))
//
//	for _, address := range addrs {
//		bs.AddressInScanning[address.Address] = sourceKey
//	}
//
//}

//IsExistAddress 指定地址是否已登记扫描
func (bs *BlockScannerBase) IsExistAddress(address string) bool {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()

	_, exist := bs.AddressInScanning[address]
	return exist
}

//IsExistWallet 指定账户的钱包是否已登记扫描
//func (bs *BlockScannerBase) IsExistWallet(accountID string) bool {
//	bs.Mu.RLock()
//	defer bs.Mu.RUnlock()
//
//	_, exist := bs.WalletInScanning[accountID]
//	return exist
//}

//AddObserver 添加观测者
func (bs *BlockScannerBase) AddObserver(obj BlockScanNotificationObject) error {
	bs.Mu.Lock()

	defer bs.Mu.Unlock()

	if obj == nil {
		return nil
	}
	if _, exist := bs.Observers[obj]; exist {
		//已存在，不重复订阅
		return nil
	}

	bs.Observers[obj] = true

	return nil
}

//RemoveObserver 移除观测者
func (bs *BlockScannerBase) RemoveObserver(obj BlockScanNotificationObject) error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()

	delete(bs.Observers, obj)

	return nil
}

//Clear 清理订阅扫描的内容
func (bs *BlockScannerBase) Clear() error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	//bs.WalletInScanning = nil
	bs.AddressInScanning = nil
	bs.AddressInScanning = make(map[string]string)
	//bs.WalletInScanning = make(map[string]*WalletWrapper)
	return nil
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *BlockScannerBase) SetRescanBlockHeight(height uint64) error {
	return nil
}

//SetTask
func (bs *BlockScannerBase) SetTask(task func()) {

	if bs.scanTask == nil {
		//创建定时器
		task := timer.NewTask(bs.PeriodOfTask, task)
		bs.scanTask = task
	}

}

//Run 运行
func (bs *BlockScannerBase) Run() error {

	if bs.Scanning {
		log.Warn("block scanner is running... ")
		return nil
	}

	if bs.scanTask == nil {
		return fmt.Errorf("block scanner has not set scan task ")
	}
	bs.Scanning = true
	bs.scanTask.Start()
	return nil
}

//Stop 停止扫描
func (bs *BlockScannerBase) Stop() error {
	bs.scanTask.Stop()
	bs.Scanning = false
	return nil
}

//Pause 暂停扫描
func (bs *BlockScannerBase) Pause() error {
	bs.scanTask.Pause()
	return nil
}

//Restart 继续扫描
func (bs *BlockScannerBase) Restart() error {
	bs.scanTask.Restart()
	return nil
}

//scanning 扫描
//func (bs *BlockScannerBase) ScanTask() {
//	//执行扫描任务
//}

//ScanBlock 扫描指定高度区块
func (bs *BlockScannerBase) ScanBlock(height uint64) error {
	//扫描指定高度区块
	return nil
}

//GetCurrentBlockHeight 获取当前区块高度
func (bs *BlockScannerBase) GetCurrentBlockHeader() (*BlockHeader, error) {
	return nil, nil
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *BlockScannerBase) GetScannedBlockHeight() uint64 {
	return 0
}

func (bs *BlockScannerBase) ExtractTransactionData(txid string) (map[string]*TxExtractData, error) {
	return nil, nil
}

//GetWalletByAddress 获取地址对应的钱包
//func (bs *BlockScannerBase) GetWalletWrapperByAddress(address string) (*WalletWrapper, bool) {
//	bs.Mu.RLock()
//	defer bs.Mu.RUnlock()
//
//	account, ok := bs.AddressInScanning[address]
//	if ok {
//		wallet, ok := bs.WalletInScanning[account]
//		return wallet, ok
//
//	} else {
//		return nil, false
//	}
//}