// Copyright (c) 2023-2024 The UXUY Developer Team
// License:
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE

package storage

import (
	"errors"
	"fmt"
	"github.com/uxuycom/indexer/model"
	"github.com/uxuycom/indexer/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math/big"
	"strings"
)

const (
	DatabaseTypeSqlite3 = "sqlite3"
	DatabaseTypeMysql   = "mysql"
)

type DBClient struct {
	SqlDB *gorm.DB
}

// NewDbClient creates a new database client instance.
func NewDbClient(cfg *utils.DatabaseConfig) (*DBClient, error) {
	gormCfg := &gorm.Config{}
	if cfg.EnableLog {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}
	switch cfg.Type {
	case DatabaseTypeSqlite3:
		return NewSqliteClient(cfg, gormCfg)
	case DatabaseTypeMysql:
		return NewMysqlClient(cfg, gormCfg)
	}
	return nil, nil
}

func (conn *DBClient) SaveLastBlock(tx *gorm.DB, status *model.BlockStatus) error {
	if tx == nil {
		return errors.New("gorm db is not valid")
	}
	return tx.Where("chain = ?", status.Chain).Save(status).Error
}

func (conn *DBClient) LastBlock(chain string) (*big.Int, error) {
	var blockNumberStr string
	err := conn.SqlDB.Raw("SELECT block_number FROM block  where `chain` = ? ORDER BY block_number DESC LIMIT 1", chain).Scan(&blockNumberStr).Error
	if err != nil {
		return nil, err
	}

	blockNumber, _ := utils.ConvetStr(blockNumberStr)
	return blockNumber, nil
}

func (conn *DBClient) BatchAddInscription(dbTx *gorm.DB, ins []*model.Inscriptions) error {
	if len(ins) < 1 {
		return nil
	}
	return dbTx.Create(ins).Error
}

func (conn *DBClient) BatchUpdateInscription(dbTx *gorm.DB, chain string, items []*model.Inscriptions) error {
	if len(items) < 1 {
		return nil
	}
	fields := map[string]string{
		"transfer_type": "%d",
	}

	vals := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		vals = append(vals, map[string]interface{}{
			"sid":           item.SID,
			"transfer_type": item.TransferType,
		})
	}
	err, _ := conn.BatchUpdatesBySID(dbTx, chain, model.Inscriptions{}.TableName(), fields, vals)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DBClient) BatchUpdatesBySID(dbTx *gorm.DB, chain string, tblName string, fields map[string]string, values []map[string]interface{}) (error, int64) {
	if len(values) < 1 {
		return nil, 0
	}

	updates := make([]string, 0, len(fields))
	for field, vt := range fields {
		update := fmt.Sprintf(" %s = CASE sid ", field)
		tpl := fmt.Sprintf(" WHEN %s THEN '%s'", "%d", vt)
		for _, value := range values {
			update += fmt.Sprintf(tpl, value["sid"], value[field])
		}
		update += " END"
		updates = append(updates, update)
	}

	ids := make([]string, 0, len(values))
	for _, value := range values {
		ids = append(ids, fmt.Sprintf("%d", value["sid"]))
	}

	finalSql := fmt.Sprintf("UPDATE %s SET %s WHERE chain = '%s' AND sid IN (%s)", tblName, strings.Join(updates, ","), chain, strings.Join(ids, ","))
	ret := dbTx.Exec(finalSql)
	if ret.Error != nil {
		return ret.Error, 0
	}
	return nil, ret.RowsAffected
}

func (conn *DBClient) BatchUpdateInscriptionStats(dbTx *gorm.DB, chain string, items []*model.InscriptionsStats) error {
	if len(items) < 1 {
		return nil
	}

	fields := map[string]string{
		"minted":  "%s",
		"holders": "%d",
		"tx_cnt":  "%d",
	}

	vals := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		vals = append(vals, map[string]interface{}{
			"sid":     item.SID,
			"minted":  item.Minted,
			"holders": item.Holders,
			"tx_cnt":  item.TxCnt,
		})
	}
	err, _ := conn.BatchUpdatesBySID(dbTx, chain, model.InscriptionsStats{}.TableName(), fields, vals)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DBClient) BatchAddInscriptionStats(dbTx *gorm.DB, ins []*model.InscriptionsStats) error {
	if len(ins) < 1 {
		return nil
	}
	return dbTx.Create(ins).Error
}

func (conn *DBClient) BatchAddTransaction(dbTx *gorm.DB, txs []*model.Transaction) error {
	if len(txs) < 1 {
		return nil
	}
	return dbTx.Create(txs).Error
}

func (conn *DBClient) BatchAddBalanceTx(dbTx *gorm.DB, txs []*model.BalanceTxn) error {
	if len(txs) < 1 {
		return nil
	}
	return dbTx.Create(txs).Error
}

func (conn *DBClient) BatchAddAddressTx(dbTx *gorm.DB, txs []*model.AddressTxs) error {
	if len(txs) < 1 {
		return nil
	}
	return dbTx.Create(txs).Error
}

func (conn *DBClient) BatchAddBalances(dbTx *gorm.DB, items []*model.Balances) error {
	if len(items) < 1 {
		return nil
	}
	return dbTx.Create(items).Error
}

func (conn *DBClient) BatchUpdateBalances(dbTx *gorm.DB, chain string, items []*model.Balances) error {
	if len(items) < 1 {
		return nil
	}

	fields := map[string]string{
		"available": "%s",
		"balance":   "%s",
	}

	vals := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		vals = append(vals, map[string]interface{}{
			"sid":       item.SID,
			"available": item.Available,
			"balance":   item.Balance,
		})
	}
	err, _ := conn.BatchUpdatesBySID(dbTx, chain, model.Balances{}.TableName(), fields, vals)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DBClient) UpdateInscriptionsStatsBySID(dbTx *gorm.DB, chain string, id uint32, updates map[string]interface{}) error {
	return dbTx.Table(model.InscriptionsStats{}.TableName()).Where("chain = ?", chain).Where("sid = ?", id).Updates(updates).Error
}

func (conn *DBClient) UpdateInscriptStatsForMint(dbTx *gorm.DB, stats *model.InscriptionsStats) error {
	ins := &model.InscriptionsStats{}
	tableName := ins.TableName()
	updateSql := ""
	var updateData []interface{}
	if stats.Minted.Sign() > 0 {
		updateSql += " minted= ? "
		updateData = append(updateData, stats.Minted)
	}

	if stats.MintCompletedTime != nil && stats.MintCompletedTime.Unix() > 0 {
		updateSql += ", mint_completed_time=? "
		updateData = append(updateData, stats.MintCompletedTime)
	}

	if stats.MintFirstBlock > 0 {
		updateSql += ", mint_first_block=? "
		updateData = append(updateData, stats.MintFirstBlock)
	}

	if stats.MintLastBlock > 0 {
		updateSql += ", mint_last_block=? "
		updateData = append(updateData, stats.MintLastBlock)
	}

	if stats.Holders > 0 {
		updateSql += ", holders=? "
		updateData = append(updateData, stats.Holders)
	}

	if stats.TxCnt > 0 {
		updateSql += ", tx_cnt=? "
		updateData = append(updateData, stats.TxCnt)
	}

	updateSql = strings.Trim(updateSql, ",")
	if len(updateSql) > 0 && len(updateData) > 0 {
		updateSql = "UPDATE " + tableName + " SET " + updateSql + "WHERE chain=? ANd protocol=? AND tick=?"
		updateData = append(updateData, stats.Chain, stats.Protocol, stats.Tick)
		err := dbTx.Exec(updateSql, updateData...).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// FindInscriptionByTick find token by tick
func (conn *DBClient) FindInscriptionByTick(chain, protocol, tick string) (*model.Inscriptions, error) {
	inscriptionBaseInfo := &model.Inscriptions{}
	err := conn.SqlDB.First(inscriptionBaseInfo, "chain = ? AND protocol = ? AND tick = ?", chain, protocol, tick).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return inscriptionBaseInfo, nil
}

// FindInscriptionStatsInfoByBaseId find inscription stats info by base id
func (conn *DBClient) FindInscriptionStatsInfoByBaseId(insId uint32) (*model.InscriptionsStats, error) {
	inscriptionStats := &model.InscriptionsStats{}
	err := conn.SqlDB.First(inscriptionStats, "ins_id = ?", insId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return inscriptionStats, nil
}

func (conn *DBClient) FindUserBalanceByTick(chain, protocol, tick, addr string) (*model.Balances, error) {
	balance := &model.Balances{}
	err := conn.SqlDB.First(balance, "chain = ? AND protocol = ? AND tick = ? AND address = ?", chain, protocol, tick, addr).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return balance, nil
}

func (conn *DBClient) FindTransaction(chain string, hash string) (*model.Transaction, error) {
	txn := &model.Transaction{}
	err := conn.SqlDB.First(txn, "chain = ? AND tx_hash = ?", chain, hash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return txn, nil
}

func (conn *DBClient) GetInscriptions(limit, offset int, chain, protocol, tick, deployBy string, sort int) (
	[]*model.InscriptionOverView, int64, error) {

	var data []*model.InscriptionOverView
	var total int64

	query := conn.SqlDB.Select("*, (d.minted / a.total_supply) as progress").Table("inscriptions as a").
		Joins("left join `inscriptions_stats` as d on (`a`.chain = `d`.chain and `a`.protocol = `d`.protocol and `a`.tick = `d`.tick)")
	if chain != "" {
		query = query.Where("`a`.chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`a`.protocol = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`a`.tick = ?", tick)
	}
	if deployBy != "" {
		query = query.Where("`a`.deploy_by = ?", deployBy)
	}

	// sort by  0.id  1.deploy_time  2.progress  3.holders  4.tx_cnt
	switch sort {
	case 0:
		query = query.Order("`a`.id desc")
	case 1:
		query = query.Order("deploy_time desc")
	case 2:
		query = query.Order("progress desc")
	case 3:
		query = query.Order("holders desc")
	case 4:
		query = query.Order("tx_cnt desc")
	}

	query = query.Count(&total)
	result := query.Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) GetInscriptionsByIdLimit(start uint64, limit int) ([]model.Inscriptions, error) {
	inscriptions := make([]model.Inscriptions, 0)
	err := conn.SqlDB.Where("id > ?", start).Order("id asc").Limit(limit).Find(&inscriptions).Error
	if err != nil {
		return nil, err
	}
	return inscriptions, nil
}

func (conn *DBClient) GetInscriptionStatsByIdLimit(start uint64, limit int) ([]model.InscriptionsStats, error) {
	stats := make([]model.InscriptionsStats, 0)
	err := conn.SqlDB.Where("id > ?", start).Order("id asc").Limit(limit).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (conn *DBClient) GetInscriptionsByAddress(limit, offset int, address string) ([]*model.Balances, error) {
	balances := make([]*model.Balances, 0)

	query := conn.SqlDB.Model(&model.Inscriptions{})
	if address != "" {
		query = query.Where("`address` = ?", address)
	}

	result := query.Order("id desc").Limit(limit).Offset(offset).Find(&balances)
	if result.Error != nil {
		return nil, result.Error
	}

	return balances, nil
}

func (conn *DBClient) GetTransactionsByAddress(limit, offset int, address, chain, protocol, tick string, event int8) (
	[]*model.AddressTransaction, int64, error) {

	var data []*model.AddressTransaction
	var total int64

	query := conn.SqlDB.Select("*").Table("txs as t").
		Joins("left join `address_txs` as a on (`t`.tx_hash = `a`.tx_hash and `t`.chain = `a`.chain and `t`.protocol = `a`.protocol and `t`.tick = `a`.tick)").
		Where("`a`.address = ?", address)

	if chain != "" {
		query = query.Where("`a`.chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`a`.protocol = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`a`.tick = ?", tick)
	}
	if event > 0 {
		query = query.Where("`a`.event = ?", event)
	}

	query = query.Count(&total)
	result := query.Order("`a`.id desc").Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) GetAddressInscriptions(limit, offset int, address, chain, protocol, tick string) (
	[]*model.BalanceInscription, int64, error) {

	var data []*model.BalanceInscription
	var total int64

	query := conn.SqlDB.Select("*").Table("balances as b").
		Joins("left join `inscriptions` as a on (`b`.chain = `a`.chain and `b`.protocol = `a`.protocol and `b`.tick = `a`.tick)")

	query = query.Where("`b`.address = ? and `b`.balance > 0", address)

	if chain != "" {
		query = query.Where("`b`.chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`b`.protocol = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`b`.tick = ?", tick)
	}

	query = query.Count(&total)
	result := query.Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) GetBalancesByAddress(limit, offset int, address, chain, protocol, tick string) (
	[]*model.Balances, int64, error) {

	var balances []*model.Balances
	var total int64

	query := conn.SqlDB.Model(&model.Balances{}).Where("`address` = ?", address)
	if chain != "" {
		query = query.Where("`chain` = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`protocol` = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`tick` = ?", tick)
	}

	query = query.Count(&total)
	err := query.Order("id desc").Limit(limit).Offset(offset).Find(&balances).Error
	if err != nil {
		return nil, 0, err
	}
	return balances, total, nil
}

func (conn *DBClient) GetHoldersByTick(limit, offset int, chain, protocol, tick string) ([]*model.Balances, int64, error) {
	var holders []*model.Balances
	var total int64
	query := conn.SqlDB.Model(&model.Balances{}).
		Where("balance > 0 and chain = ? and protocol = ? and tick = ?", chain, protocol, tick)
	query = query.Count(&total)
	result := query.Order("id desc").Limit(limit).Offset(offset).Find(&holders)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return holders, total, nil
}

func (conn *DBClient) GetUTXOCount(address, chain, protocol, tick string) (int64, error) {
	var count int64
	query := conn.SqlDB.Model(&model.UTXO{}).
		Where("address = ? and chain = ? and protocol = ? and tick = ? and status = ?", address, chain, protocol, tick, model.UTXOStatusUnspent)
	err := query.Count(&count)
	if err.Error != nil {
		return 0, err.Error
	}
	return count, nil
}

func (conn *DBClient) GetBalancesByIdLimit(start uint64, limit int) ([]model.Balances, error) {
	balances := make([]model.Balances, 0)
	err := conn.SqlDB.Where("id > ?", start).Order("id asc").Limit(limit).Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return balances, nil
}

func (conn *DBClient) GetUTXOsByIdLimit(start uint64, limit int) ([]model.UTXO, error) {
	utxos := make([]model.UTXO, 0, limit)
	err := conn.SqlDB.Where("id > ? ", start).Where("status = ? ", model.UTXOStatusUnspent).Order("id asc").Limit(limit).Find(&utxos).Error
	if err != nil {
		return nil, err
	}
	return utxos, nil
}

func (conn *DBClient) FindUtxoByAddress(tx *gorm.DB, address, tick string) (*model.UTXO, error) {
	utxo := &model.UTXO{}
	err := conn.SqlDB.First(utxo, "address = ? and tick = ? ", address, tick).Error
	if err != nil {
		return nil, err
	}

	return utxo, nil
}

func (conn *DBClient) FirstValidUtxoByRootHash(tx *gorm.DB, chain, txid, address string) (*model.UTXO, error) {
	utxo := &model.UTXO{}
	err := conn.SqlDB.First(utxo, "address = ? AND root_hash = ? AND chain = ? AND status = ?", address, txid, chain, model.UTXOStatusUnspent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return utxo, nil
}

func (conn *DBClient) FirstUTXOByRootHash(tx *gorm.DB, chain, txid string) (*model.UTXO, error) {
	utxo := &model.UTXO{}
	err := conn.SqlDB.First(utxo, " root_hash = ? AND chain = ?", txid, chain).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return utxo, nil
}

func (conn *DBClient) GetUtxosByAddress(address, chain, protocol, tick string) ([]*model.UTXO, error) {
	var utxos []*model.UTXO
	query := conn.SqlDB.Model(&model.UTXO{}).
		Where("address = ? and chain = ? and protocol = ? and tick = ? and status = ?", address, chain, protocol, tick, model.UTXOStatusUnspent)
	result := query.Order("id desc").Find(&utxos)
	if result.Error != nil {
		return nil, result.Error
	}
	return utxos, nil
}

func (conn *DBClient) FindAddressTxByHash(chain, hash string) (*model.AddressTxs, error) {
	tx := &model.AddressTxs{}
	err := conn.SqlDB.First(tx, "chain = ? and tx_hash = ? ", chain, hash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return tx, nil
}
