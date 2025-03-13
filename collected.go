
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	//"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// информация о транзакции
type Transaction struct {
	Hash     string   `json:"hash"`
	ChainId  *big.Int `json:"chain_id"`
	Value    *big.Int `json:"value"`
	To       string   `json:"to"`
	Gas      uint64   `json:"gas"`
	GasPrice *big.Int `json:"gas_price"`
}

// информация о блоке и его транзакциях
type BlockInfo struct {
	BlockNumber      uint64        `json:"block_number"`
	BlockTime        uint64        `json:"block_time"`
	BlockDifficulty  uint64        `json:"block_difficulty"`
	BlockHash        string        `json:"block_hash"`
	TransactionCount int           `json:"transaction_count"`
	Transactions     []Transaction `json:"transactions"`
	Error            error         `json:"-"` 
}

var lastBlockHash string

//  информация о последнем блоке и его транзакциях
func GetLatestBlockInfo(clientURL string) BlockInfo {
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		return BlockInfo{Error: fmt.Errorf("ошибка подключения к клиенту Ethereum: %w", err)}
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return BlockInfo{Error: fmt.Errorf("ошибка получения заголовка последнего блока: %w", err)}
	}

	blockNumber := big.NewInt(header.Number.Int64())
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return BlockInfo{Error: fmt.Errorf("ошибка получения блока: %w", err)}
	}

	var transactions []Transaction
	for _, tx := range block.Transactions() {
		var toAddress string
		if tx.To() != nil {
			toAddress = tx.To().Hex()
		}

		transactions = append(transactions, Transaction{
			Hash:     tx.Hash().Hex(),
			ChainId:  tx.ChainId(),
			Value:    tx.Value(),
			To:       toAddress,
			Gas:      tx.Gas(),
			GasPrice: tx.GasPrice(),
		})
	}

	return BlockInfo{
		BlockNumber:      block.Number().Uint64(),
		BlockTime:        block.Time(),
		BlockDifficulty:  block.Difficulty().Uint64(),
		BlockHash:        block.Hash().Hex(),
		TransactionCount: len(block.Transactions()),
		Transactions:     transactions,
	}
}

// с бд
func uploadData(block BlockInfo, firebaseURL string) error {
	blockJson, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("ошибка сериализации данных: %w", err)
	}

	req, err := http.NewRequest("POST", firebaseURL, bytes.NewBuffer(blockJson))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки данных: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("неожиданный HTTP-статус ответа: %d", resp.StatusCode)
	}

	lastBlockHash = block.BlockHash
	return nil
}

